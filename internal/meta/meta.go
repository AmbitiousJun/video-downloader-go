package meta

import (
	"fmt"
	"strings"
	"video-downloader-go/internal/util/myhttp"
	"video-downloader-go/internal/util/mylog/dlbar"

	"github.com/google/uuid"
)

// 用于分割下载链接的分割串
var YtdlLinksSep = uuid.New().String()

// 视频文件元数据
type Video struct {
	LogBar *dlbar.Bar // 日志任务条
	Name   string     // 视频名称
	Url    string     // 视频地址
}

// Download 封装了一个视频下载任务所需要的元数据
type Download struct {
	LogBar    *dlbar.Bar        // 日志任务条
	Link      string            // 视频下载地址
	FileName  string            // 视频名称
	OriginUrl string            // 源视频地址
	HeaderMap map[string]string // 请求头
}

// 创建一个适配 youtube-dl 的下载元数据
func NewYtDlDownloadMeta(links []string, fileName, originUrl string) *Download {
	return NewDownloadMeta(strings.Join(links, YtdlLinksSep), fileName, originUrl)
}

// 将一个 link 分割为 link 数组
func Split2YtDlLinks(link string) []string {
	return strings.Split(link, YtdlLinksSep)
}

// NewDownloadMeta 用于创建一个默认的下载元数据
func NewDownloadMeta(link, fileName, originUrl string) *Download {
	dm := Download{Link: link, FileName: fileName, OriginUrl: originUrl}
	m := myhttp.GenDefaultHeaderMapByUrl(nil, link)
	dm.HeaderMap = m
	return &dm
}

// 格式化输出
func (v *Video) String() string {
	return fmt.Sprintf("[Video[Name: %v, Url: %v]]", v.Name, v.Url)
}

// 格式化输出
func (d *Download) String() string {
	return fmt.Sprintf(
		"[Download[Link: %v, FileName: %v, OriginUrl: %v, HeaderMap: %v]]",
		d.Link,
		d.FileName,
		d.OriginUrl,
		d.HeaderMap,
	)
}
