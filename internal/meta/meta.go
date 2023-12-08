package meta

import "video-downloader-go/internal/util/myhttp"

// 视频文件元数据
type Video struct {
	Name string // 视频名称
	Url  string // 视频地址
}

// Download 封装了一个视频下载任务所需要的元数据
type Download struct {
	Link      string            // 视频下载地址
	FileName  string            // 视频名称
	OriginUrl string            // 源视频地址
	HeaderMap map[string]string // 请求头
}

// NewDownloadMeta 用于创建一个默认的下载元数据
func NewDownloadMeta(link, fileName, originUrl string) *Download {
	dm := Download{Link: link, FileName: fileName, OriginUrl: originUrl}
	m := myhttp.GenDefaultHeaderMapByUrl(nil, link)
	dm.HeaderMap = m
	return &dm
}
