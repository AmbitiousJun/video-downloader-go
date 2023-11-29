package coredl

import (
	"net/http"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/myhttp"

	"github.com/pkg/errors"
)

// mp4 单协程下载
type Mp4SimpleDownloader struct{}

// mp4 多协程下载
type Mp4MultiThreadDownloader struct{}

// 使用 mp4 单协程下载时，current 和 total 分别是已下载字节数和总字节数
func (d *Mp4SimpleDownloader) Exec(meta *meta.Download, progressListener func(current, total int64)) error {
	var current, total int64 = 0, 0
	// 1 获取文件总大小
	ranges, err := myhttp.GetRequestRangesFrom(meta.Link, http.MethodGet, myhttp.GenDefaultHeaderMapByUrl(nil, meta.Link), 0)
	if err != nil {
		return errors.Wrap(err, "无法获取文件总大小")
	}
	total = ranges[1]
	// 2 分片

	return nil
}
