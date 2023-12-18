// 使用命令行的方式执行下载器

package downloader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/downloader/ytdl"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/myfile"
	"video-downloader-go/internal/util/mylog"

	"golang.org/x/term"
)

// 错误信息
const (
	UnValidM3U8 = "不是规范的 m3u8 地址"
)

// 任务下载完成处理函数
type CompleteOne func()

// 任务下载失败的监听器，下载器会将失败的任务传递出来
type DlErrorHandler func(dmt *meta.Download)

// ListenAndDownload 用于命令行模式下监听下载任务并依据全局配置多协程下载任务
func ListenAndDownload(list *meta.TaskDeque[meta.Download], completeOne CompleteOne, dlErrorHandler DlErrorHandler) {
	mylog.Info("开始监听下载列表...")
	go func() {
		for {
			if list.Empty() {
				// 没有下载任务，睡眠两秒
				time.Sleep(time.Second * 2)
				continue
			}

			// 取出一个下载任务来下载
			dmt := list.PollFirst()
			handleTask(dmt, completeOne, dlErrorHandler, func(d *meta.Download) {
				list.OfferLast(dmt)
			})
		}
	}()
}

// handleTask 是处理一个下载任务，使用的是协程池中的 goroutine
func handleTask(dmt *meta.Download, completeOne CompleteOne, dlErrorHandler DlErrorHandler, offerBack func(*meta.Download)) {

	dlpool.SubmitTask(func() {
		originFilename := dmt.FileName
		link := dmt.Link
		fileName := fmt.Sprintf("%s%s%s.mp4", config.G.Downloader.DownloadDir, string(filepath.Separator), originFilename)
		dmt.FileName = fileName
		mylog.Infof("监听到下载任务，文件名：%v，下载地址：%v", fileName, link)

		// 初始化下载器并下载
		cdl := initCoreDownloader()
		err := cdl.Exec(dmt, func(p *coredl.Progress) {
			printDownloadProgress(dmt.FileName, p)
		})

		// 下载成功
		if err == nil {
			completeOne()
			return
		}

		// 下载出现异常，检查是否有下载一半的文件，将其删除
		myfile.DeleteAnyFileContainsPrefix(dmt.FileName)

		// 下载失败，无效的 m3u8
		if err.Error() == UnValidM3U8 {
			dmt.FileName = originFilename
			mylog.Warnf("下载失败：%v, 重新添加到解析任务中，视频名称：%v", err, dmt.FileName)
			// 触发下载异常
			dlErrorHandler(dmt)
			return
		}

		// 其他下载异常
		mylog.Errorf("下载失败：%v，重新加入下载列表", err)
		offerBack(dmt)
	})
}

// printDownloadProgress 负责将下载进度日志输出到控制台上
func printDownloadProgress(fileName string, p *coredl.Progress) {
	percent := float64(p.Current) / float64(p.Total)
	curDlMb := float64(p.CurrentBytes) / 1024 / 1024
	totDlMb := float64(p.TotalBytes) / 1024 / 1024

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
		// 清空最后 9 行日志
		mylog.Success("\033[9F\033[J\r")
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
	mylog.Successf("分片进度：%v/%v(%.2f%%)", p.Current, p.Total, percent*100)
	mylog.Successf("文件大小：%.2f/%.2f (MB)", curDlMb, totDlMb)
	mylog.Successf("任务进度：%v/%v", p.CurrentTask, p.TotalTasks)
	mylog.Successf("下载速率：%s", config.RateLimitBucket().CurrentRateStr)
	mylog.Success("==== 下载进度 ⬆️")
}

// coreDownloaderCache 用于缓存初始化好的下载器
// 在命令行模式下，只能通过全局配置文件初始化一个下载器
var coreDownloaderCache coredl.Downloader

// initDownloaderOnce 用于控制下载器初始化逻辑只执行一次
var initDownloaderOnce sync.Once

// initCoreDownloader 根据全局配置初始化一个下载器对象
// 如果初始化失败，程序将直接退出
func initCoreDownloader() coredl.Downloader {

	initDownloaderOnce.Do(func() {
		// 如果是通过 youtube-dl 解析的，就使用适配的下载器
		if config.G.Decoder.Use == config.DecoderYoutubeDl {
			coreDownloaderCache = ytdl.New()
			return
		}

		// 获取配置
		resource := config.G.Decoder.ResourceType
		dlType := config.G.Downloader.Use

		// 生成对象
		switch resource + dlType {

		case config.ResourceMP4 + config.DownloadSimple:
			coreDownloaderCache = coredl.NewMp4Simple()

		case config.ResourceMP4 + config.DownloadMultiThread:
			coreDownloaderCache = coredl.NewMp4MultiThread()

		case config.ResourceM3U8 + config.DownloadSimple:
			coreDownloaderCache = coredl.NewM3U8Simple()

		case config.ResourceM3U8 + config.DownloadMultiThread:
			coreDownloaderCache = coredl.NewM3U8MultiThread()

		default:
			log.Fatal("下载器初始化异常，请检查配置")
		}
	})

	return coreDownloaderCache
}
