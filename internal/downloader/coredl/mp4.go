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

// 使用 mp4 单协程下载时，current 和 total 分别是已下载字节数和总字节数
func (d *mp4SimpleDownloader) Exec(dmt *meta.Download, progressListener func(current, total int64)) error {
	return downloadMp4(dmt, progressListener, false)
}

func (d *mp4MultiThreadDownloader) Exec(dmt *meta.Download, progressListener func(current, total int64)) error {
	return downloadMp4(dmt, progressListener, true)
}

// downloadMp4 函数定义了核心的下载逻辑
func downloadMp4(dmt *meta.Download, progressListener func(current, total int64), multiThread bool) (err error) {
	var taskCount int
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("下载时出现异常：%v, 请检查各项配置是否正确", r))
			appctx.BatchDone(taskCount)
		}
	}()
	var current, total int64 = 0, 0
	// 1 获取文件总大小
	ranges, err := myhttp.GetRequestRangesFrom(dmt.Link, http.MethodGet, myhttp.GenDefaultHeaderMapByUrl(nil, dmt.Link), 0)
	if err != nil || ranges[1] <= 0 {
		return errors.Wrap(err, "无法获取文件总大小或文件为空")
	}
	total = ranges[1]
	// 调用一次监听器，使得调用方可以获得文件的总大小
	progressListener(current, total)
	// 2 分片
	tasks := initUnitTasks(total)
	taskCount = len(tasks)
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
		// 使用主函数的 err 对象传递错误信息
		var newReq *http.Request
		newReq, err = myhttp.CloneHttpRequest(req)
		if err != nil {
			err = errors.Wrapf(err, "克隆请求时出现异常：%v", dmt)
			return
		}
		newReq.Header.Set(myhttp.HttpHeaderRangesKey, fmt.Sprintf("bytes=%d-%d", task.from, task.to))
		if err = myhttp.DownloadWithRateLimit(newReq, dmt.FileName); err != nil {
			err = errors.Wrapf(err, "下载分片时出现异常：%v, %v", dmt, task)
			return
		}
		// 每下载完成一个分片就通知一次监听器
		atomic.AddInt64(&current, task.to-task.from)
		progressListener(current, total)
	}
	// 定义一个局部的协程同步器，为了把当前函数变为同步的
	var wg sync.WaitGroup
	wg.Add(taskCount)
	appctx.WaitGroup().Add(taskCount)
	for _, task := range tasks {
		if !multiThread {
			if downloadTask(task); err != nil {
				return
			}
			wg.Done()
			appctx.WaitGroup().Done()
			continue
		}
		err = dlpool.Download.Submit(func() {
			defer appctx.WaitGroup().Done()
			defer wg.Done()
			select {
			case <-appctx.Context().Done():
				return
			default:
				if err != nil {
					// 如果下载已经出错了，就没必要再下了
					return
				}
				downloadTask(task)
			}
		})
		if err != nil {
			panic(errors.Wrap(err, "协程池运行异常"))
		}
	}
	// 等待所有的协程运行完毕
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
