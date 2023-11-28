// 负责多线程处理下载任务

package dlthreadpool

import "video-downloader-go/internal/config"

type Pool struct {
	PoolChannel chan struct{}  // 用于同步协程的通道
	PoolSize    int            // 线程池大小
	WaitQueue   []func() error // 等待队列
}

// 处理任务的线程池
var TaskPool *Pool

// 处理下载任务的线程池
var DownloadPool *Pool

// 从全局配置中初始化线程池
func InitFromGlobalConfig() {
	tasks := config.GlobalConfig.Downloader.TaskThreadCount
	downloads := config.GlobalConfig.Downloader.DlThreadCount
	TaskPool = &Pool{
		PoolSize:    tasks,
		WaitQueue:   []func() error{},
		PoolChannel: make(chan struct{}, tasks),
	}
	DownloadPool = &Pool{
		PoolSize:    downloads,
		WaitQueue:   []func() error{},
		PoolChannel: make(chan struct{}, downloads),
	}
}
