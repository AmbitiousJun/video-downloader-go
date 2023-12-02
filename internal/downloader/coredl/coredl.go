// 把 mp4, m3u8 两种下载格式的下载处理函数封装在 coredl 包中

package coredl

import "video-downloader-go/internal/meta"

type Downloader interface {
	// Exec 是下载器的核心处理方法
	Exec(dmt *meta.Download, handlerFunc processHandler) error
}

// processHandler 可以使调用方实时获取下载进度
type processHandler func(current, total int64)

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
