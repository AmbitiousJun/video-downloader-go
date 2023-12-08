// 负责多协程处理下载任务

package dlpool

import (
	"sync"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
)

// 处理任务的线程池
var task *ants.Pool

// 处理下载任务的协程池
var download *ants.Pool

// 用于控制单次初始化
var once sync.Once

// SubmitDownload 用于异步提交下载任务
func SubmitDownload(d func()) error {
	return submit(download, d)
}

// SubmitTask 用于异步提交视频下载任务
func SubmitTask(t func()) error {
	return submit(task, t)
}

func submit(p *ants.Pool, handleFunc func()) error {
	if err := initOnce(); err != nil {
		return err
	}
	return task.Submit(func() {
		select {
		case <-appctx.Context().Done():
			// 释放资源并退出
			releaseAll()
			return
		default:
			handleFunc()
		}
	})
}

// initOnce 用于全局初始化一次协程次配置
func initOnce() (err error) {
	once.Do(func() {
		err = initFromGlobalConfig()
	})
	return
}

// 从全局配置中初始化协程池
func initFromGlobalConfig() (err error) {
	defer func() {
		if err != nil {
			// 清除初始化一半的协程池
			releaseAll()
		}
	}()
	tasks := config.GlobalConfig.Downloader.TaskThreadCount
	downloads := config.GlobalConfig.Downloader.DlThreadCount
	task, err = ants.NewPool(tasks, ants.WithOptions(commonAntsOptions()))
	if err != nil {
		err = errors.Wrap(err, "初始化下载任务协程池失败")
		return
	}
	download, err = ants.NewPool(downloads, ants.WithOptions(commonAntsOptions()))
	if err != nil {
		err = errors.Wrap(err, "初始化下载协程池失败")
		return
	}
	return
}

// 防止多次释放资源
var releaseMutex sync.Mutex

// 销毁所有的协程池
func releaseAll() {
	releaseMutex.Lock()
	defer releaseMutex.Unlock()
	if task != nil {
		task.Release()
	}
	if download != nil {
		download.Release()
	}
}

func commonAntsOptions() ants.Options {
	return ants.Options{
		ExpiryDuration:   time.Minute * 10, // worker 10 分钟没有任务处理就销毁
		MaxBlockingTasks: 1024 * 1024,      // 最多可以有多少阻塞的任务
	}
}
