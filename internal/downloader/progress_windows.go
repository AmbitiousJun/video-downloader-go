//go:build windows
// +build windows

package downloader

import (
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/util/mylog"
)

// printDownloadProgress 负责将下载进度日志输出到控制台上
func printDownloadProgress(fileName string, p *coredl.Progress) {
	percent := float64(p.Current) / float64(p.Total)
	curDlMb := float64(p.CurrentBytes) / 1024 / 1024
	totDlMb := float64(p.TotalBytes) / 1024 / 1024

	mylog.Success("==== 下载进度 ⬇️")
	mylog.Successf("文件名：%v", fileName)
	mylog.Successf("分片进度：%v/%v(%.2f%%)", p.Current, p.Total, percent*100)
	mylog.Successf("文件大小：%.2f/%.2f (MB)", curDlMb, totDlMb)
	mylog.Successf("任务进度：%v/%v", p.CurrentTask, p.TotalTasks)
	mylog.Successf("下载速率：%s", config.RateLimitBucket().CurrentRateStr)
	mylog.Success("==== 下载进度 ⬆️")
}
