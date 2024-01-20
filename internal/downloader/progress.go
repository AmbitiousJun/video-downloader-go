package downloader

import (
	"os"
	"strings"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/util/mylog"

	"golang.org/x/term"
)

// calcDownloadSize 计算当前下载文件的大小
//
// 当下载大小 < 1024 MB 时，展示单位是 MB
//
// 当下载大小 >= 1024 MB 时，展示单位是 GB
func calcDownloadSize(p *coredl.Progress) (float64, string, float64, string) {
	curDl := float64(p.CurrentBytes) / 1024 / 1024
	totDl := float64(p.TotalBytes) / 1024 / 1024
	curSuffix, totSuffix := "MB", "MB"

	if curDl >= 1024 {
		curDl /= 1024
		curSuffix = "GB"
	}

	if totDl >= 1024 {
		totDl /= 1024
		totSuffix = "GB"
	}

	return curDl, curSuffix, totDl, totSuffix
}

// printDownloadProgress 负责将下载进度日志输出到控制台上
func printDownloadProgress(fileName string, p *coredl.Progress) {
	percent := float64(p.Current) / float64(p.Total)
	curDl, curSuffix, totDl, totSuffix := calcDownloadSize(p)

	// 获取控制台中一行可以显示多少个字符，用于显示进度条
	fd := int(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)

	var totalBlocks, finishBlocks int
	if err == nil {
		// 28 是日志输出前缀长度：'2023/12/17 01:15:15 SUCCESS '
		// 2 是输出进度条的时候的左右括号
		totalBlocks = width - 28 - 2
		finishBlocks = int(float64(totalBlocks) * percent)
		if p.Current == p.Total-1 {
			// 剩最后一个分片时，进度条拉满
			finishBlocks = totalBlocks
		}
	}

	if err == nil {
		if p.Current == 0 {
			// 首次输出时多输出一个空行
			mylog.Success("")
		}

		if p.Current != 0 {
			// 清空最后 9 行日志
			mylog.Success("\033[9F\033[J\r")
		}

		mylog.Success("==== 下载进度 ⬇️")

		// 输出进度条
		mylog.Successf("[%v%v]", strings.Repeat("*", finishBlocks), strings.Repeat("-", totalBlocks-finishBlocks))

		// 控制文件名的长度不超过 1 行
		maxLen := int(float64(width) * 0.6)
		if len(fileName) > maxLen {
			fileName = fileName[:maxLen] + "..."
		}
	} else {
		mylog.Success("==== 下载进度 ⬇️")
		mylog.Success("")
	}

	mylog.Successf("文件名：%v", fileName)
	mylog.Successf("分片进度：%v/%v (%.2f%%)", p.Current, p.Total, percent*100)
	mylog.Successf("文件大小：%.2f (%s) / %.2f (%s)", curDl, curSuffix, totDl, totSuffix)
	mylog.Successf("任务进度：%v/%v", p.CurrentTask, p.TotalTasks)
	mylog.Successf("下载速率：%s", config.RateLimitBucket().CurrentRateStr)
	mylog.Success("==== 下载进度 ⬆️")
}
