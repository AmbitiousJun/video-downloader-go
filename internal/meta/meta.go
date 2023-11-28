package entity

import "video-downloader-go/internal/util/myhttp"

// 视频文件元数据
type Video struct {
	Name string // 视频名称
	Url  string // 视频地址
}

type Download struct {
	Link      string            // 视频下载地址
	FileName  string            // 视频名称
	OriginUrl string            // 源视频地址
	HeaderMap map[string]string // 请求头
}

func NewDownloadMeta(link, fileName, originUrl string) *Download {
	dm := Download{Link: link, FileName: fileName, OriginUrl: originUrl}
	m := myhttp.GenDefaultHeaderMapByUrl(nil, link)
	dm.HeaderMap = m
	return &dm
}