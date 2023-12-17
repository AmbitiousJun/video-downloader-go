//go:build !windows
// +build !windows

package downloader

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/util/mylog"
)

// printDownloadProgress 负责将下载进度日志输出到控制台上
func printDownloadProgress(fileName string, p *coredl.Progress) {
	percent := float64(p.Current) / float64(p.Total)
	curDlMb := float64(p.CurrentBytes) / 1024 / 1024
	totDlMb := float64(p.TotalBytes) / 1024 / 1024

	// 获取控制台中一行可以显示多少个字符，用于显示进度条
	var dimensions [4]uint16
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0)

	var totalBlocks, finishBlocks int
	if err == 0 {
		// 28 是日志输出前缀长度：'2023/12/17 01:15:15 SUCCESS '
		// 2 是输出进度条的时候的左右括号
		totalBlocks = int(dimensions[1] - 28 - 2)
		finishBlocks = int(float64(totalBlocks) * percent)
		if p.Current == p.Total-1 {
			// 剩最后一个分片时，进度条拉满
			finishBlocks = totalBlocks
		}
	}

	if err == 0 {
		// 清空最后 9 行日志
		mylog.Success("\033[9F\033[J\r")
		mylog.Success("==== 下载进度 ⬇️")

		// 输出进度条
		mylog.Successf("[%v%v]", strings.Repeat("*", finishBlocks), strings.Repeat("-", totalBlocks-finishBlocks))

		// 控制文件名的长度不超过 1 行
		maxLen := int(float64(dimensions[1]) * 0.6)
		if len(fileName) > maxLen {
			fileName = fileName[:maxLen] + "..."
		}
	} else {
		mylog.Success("==== 下载进度 ⬇️")
		mylog.Success("")
	}

	mylog.Successf("文件名：%v", fileName)
	mylog.Successf("分片进度：%v/%v(%.2f%%)", p.Current, p.Total, percent*100)
	mylog.Successf("文件大小：%.2f/%.2f (MB)", curDlMb, totDlMb)
	mylog.Successf("任务进度：%v/%v", p.CurrentTask, p.TotalTasks)
	mylog.Successf("下载速率：%s", config.RateLimitBucket().CurrentRateStr)
	mylog.Success("==== 下载进度 ⬆️")
}
