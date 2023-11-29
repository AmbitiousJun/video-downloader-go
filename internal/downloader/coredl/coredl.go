// 把 mp4, m3u8 两种下载格式的下载处理函数封装在 coredl 包中

package coredl

import (
	"video-downloader-go/internal/meta"
)

type Downloader interface {
	// Exec 是下载器的核心处理函数，传入下载元数据和一个进度监听器进行下载
	Exec(meta *meta.Download, progressListener func(current, total int64)) error
}
