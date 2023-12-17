// 把 mp4, m3u8 两种下载格式的下载处理函数封装在 coredl 包中

package coredl

import "video-downloader-go/internal/meta"

type Downloader interface {
	// Exec 是下载器的核心处理方法
	Exec(dmt *meta.Download, handlerFunc ProgressHandler) error
}

// Progress 是下载进度记录结构
type Progress struct {
	Current int64 // 当前下载进度（已下载的分片数）
	Total   int64 // 任务总大小（总分片数）

	CurrentBytes int64 // 当前下载的字节数
	TotalBytes   int64 // 总字节数（下载 m3u8 时，这个值与 CurrentBytes 保持一致）

	CurrentTask int // 当前正在执行第几个任务
	TotalTasks  int // 总共需要执行的任务数
}

// ProgressHandler 可以使调用方实时获取下载进度
type ProgressHandler func(progress *Progress)

// 初始化一个 m3u8 单协程下载器
func NewM3U8Simple() Downloader {
	return new(m3u8SimpleDownloader)
}

// 初始化一个 m3u8 多协程下载器
func NewM3U8MultiThread() Downloader {
	return new(m3u8MultiThreadDownloader)
}

// 初始化一个 mp4 单协程下载器
func NewMp4Simple() Downloader {
	return new(mp4SimpleDownloader)
}

// 初始化一个 mp4 多协程下载器
func NewMp4MultiThread() Downloader {
	return new(mp4MultiThreadDownloader)
}
