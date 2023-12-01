// 负责多协程处理下载任务

package dlpool

import (
	"time"
	"video-downloader-go/internal/config"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"
)

// 处理任务的线程池
var Task *ants.Pool

// 处理下载任务的协程池
var Download *ants.Pool

// 从全局配置中初始化协程池
func InitFromGlobalConfig() (err error) {
	defer func() {
		if err != nil {
			// 清除初始化一半的协程池
			ReleaseAll()
		}
	}()
	tasks := config.GlobalConfig.Downloader.TaskThreadCount
	downloads := config.GlobalConfig.Downloader.DlThreadCount
	Task, err = ants.NewPool(tasks, ants.WithOptions(commonAntsOptions()))
	if err != nil {
		err = errors.Wrap(err, "初始化下载任务协程池失败")
		return
	}
	Download, err = ants.NewPool(downloads, ants.WithOptions(commonAntsOptions()))
	if err != nil {
		err = errors.Wrap(err, "初始化下载协程池失败")
		return
	}
	return
}

// 销毁所有的协程池
func ReleaseAll() {
	if Task != nil {
		Task.Release()
	}
	if Download != nil {
		Download.Release()
	}
}

func commonAntsOptions() ants.Options {
	return ants.Options{
		ExpiryDuration:   time.Minute * 10, // worker 10 分钟没有任务处理就销毁
		MaxBlockingTasks: 1024 * 1024,      // 最多可以有多少阻塞的任务
	}
}
