// mp4 视频下载
package coredl

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util"
	"video-downloader-go/internal/util/myhttp"
	"video-downloader-go/internal/util/mymath"

	"github.com/pkg/errors"
)

const (
	SplitCount = 64 // 一个 MP4 文件至少分割成多少份
)

// mp4 单协程下载
type mp4SimpleDownloader struct{}

// mp4 多协程下载
type mp4MultiThreadDownloader struct{}

// 一个 mp4 任务分割出来的分片任务
type unitTask struct {
	from int64 // 起始字节（闭）
	to   int64 // 终止字节（开）
}

func (d *mp4SimpleDownloader) Exec(dmt *meta.Download, handlerFunc processHandler) error {
	return downloadMp4(dmt, handlerFunc, false)
}

func (d *mp4MultiThreadDownloader) Exec(dmt *meta.Download, handlerFunc processHandler) error {
	return downloadMp4(dmt, handlerFunc, true)
}

// downloadMp4 函数定义了核心的下载逻辑
func downloadMp4(dmt *meta.Download, handlerFunc processHandler, multiThread bool) (err error) {
	var current, total int64 = 0, 0
	// 1 获取文件总大小
	ranges, err := myhttp.GetRequestRangesFrom(dmt.Link, http.MethodGet, myhttp.GenDefaultHeaderMapByUrl(nil, dmt.Link), 0)
	if err != nil {
		if util.IsRetryableError(err) {
			time.Sleep(time.Second * 2)
			return downloadMp4(dmt, handlerFunc, multiThread)
		}
		return errors.Wrap(err, "无法获取文件总大小")
	}
	total = ranges[1]
	if total <= 0 {
		return errors.New("空文件，停止下载")
	}
	// 调用一次监听器，使得调用方可以获得文件的总大小
	handlerFunc(current, total)
	// 2 分片
	tasks := initUnitTasks(total)
	// 3 循环分片进行下载
	defaultHeaders := myhttp.GenDefaultHeaderMapByUrl(nil, dmt.Link)
	// 构造请求，携带上分片头
	req, err := http.NewRequest(http.MethodGet, dmt.Link, nil)
	if err != nil {
		return errors.Wrapf(err, "构造请求时出现异常：%v", dmt)
	}
	for k, v := range defaultHeaders {
		req.Header.Add(k, v)
	}
	downloadTask := func(task *unitTask) {
		var tmpErr error
		defer func() {
			err = util.AnyError(err, tmpErr)
		}()
		// 使用主函数的 err 对象传递错误信息
		var newReq *http.Request
		newReq, tmpErr = myhttp.CloneHttpRequest(req)
		if tmpErr != nil {
			tmpErr = errors.Wrapf(tmpErr, "克隆请求时出现异常：%v", dmt)
			return
		}
		newReq.Header.Set(myhttp.HttpHeaderRangesKey, fmt.Sprintf("bytes=%d-%d", task.from, task.to))
		if tmpErr = myhttp.DownloadWithRateLimit(newReq, dmt.FileName); tmpErr != nil {
			tmpErr = errors.Wrapf(tmpErr, "下载分片时出现异常：%v, %v", dmt, task)
			return
		}
		// 每下载完成一个分片就通知一次监听器
		atomic.AddInt64(&current, task.to-task.from)
		handlerFunc(current, total)
	}
	if multiThread {
		err = util.AnyError(err, handleTasksMultiThread(tasks, downloadTask))
	} else {
		handleTasksSimple(tasks, downloadTask)
	}
	return
}

// 单协程处理任务列表
func handleTasksSimple(tasks []*unitTask, downloadTaskFunc func(*unitTask)) {
	if len(tasks) == 0 {
		return
	}
	for _, task := range tasks {
		downloadTaskFunc(task)
	}
}

// 多协程处理任务列表
func handleTasksMultiThread(tasks []*unitTask, downloadTaskFunc func(*unitTask)) (err error) {
	if len(tasks) == 0 {
		return
	}
	taskCount := len(tasks)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("下载时出现异常：%v, 请检查各项配置是否正确", r))
			appctx.BatchDone(taskCount)
		}
	}()
	// 协程同步器，下载是多协程下载，但是函数仍然是同步执行完成的
	var wg sync.WaitGroup
	wg.Add(taskCount)
	appctx.WaitGroup().Add(taskCount)
	for _, task := range tasks {
		// 在 for-range 结构中使用多协程时需要拷贝指针
		copyTask := task
		err = dlpool.Download.Submit(func() {
			defer wg.Done()
			defer appctx.WaitGroup().Done()
			select {
			case <-appctx.Context().Done():
				return
			default:
				if err == nil {
					downloadTaskFunc(copyTask)
				}
			}
		})
		if err != nil {
			panic(errors.Wrap(err, "协程池运行异常"))
		}
	}
	// 阻塞等待所有协程运行完毕
	wg.Wait()
	return
}

// 初始化下载任务分片列表
func initUnitTasks(fileTotalSize int64) []*unitTask {
	size := int64(math.Ceil(float64(fileTotalSize) / SplitCount))
	// 每个分片大小 2~4 MB
	var baseSize int64 = 2 * 1024 * 1024
	var i int64
	res := []*unitTask{}
	for ; i < SplitCount; i++ {
		curSize := int64(mymath.Min(size, fileTotalSize-i*size))
		from := i * curSize
		rd := rand.New(rand.NewSource(time.Now().UnixNano()))
		for curSize > 2*baseSize {
			random := int64(float64(baseSize) * rd.Float64())
			to := from + baseSize + random
			res = append(res, &unitTask{from: from, to: to})
			curSize -= (baseSize + random)
			from = to
		}
		if curSize > 0 {
			res = append(res, &unitTask{from: from, to: from + curSize})
		}
	}
	return res
}
