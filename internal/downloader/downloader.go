// 使用命令行的方式执行下载器

package downloader

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
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
		mylog.Info(fmt.Sprintf("监听到下载任务，文件名：%v，下载地址：%v", fileName, link))

		// 初始化下载器并下载
		cdl := initCoreDownloader()
		err := cdl.Exec(dmt, func(p *coredl.Progress) {
			percent := float64(p.Current) / float64(p.Total) * 100
			curDlMb := float64(p.CurrentBytes) / 1024 / 1024
			totDlMb := float64(p.TotalBytes) / 1024 / 1024
			mylog.Success(fmt.Sprintf("当前文件下载进度：%v/%v(%.2f%%)，已下载：%.2f MB，总共：%.2f MB", p.Current, p.Total, percent, curDlMb, totDlMb))
		})

		// 下载成功
		if err == nil {
			completeOne()
			return
		}

		// 下载失败，无效的 m3u8
		if strings.EqualFold(err.Error(), UnValidM3U8) {
			dmt.FileName = originFilename
			mylog.Warn(fmt.Sprintf("下载失败：%v, 重新添加到解析任务中，视频名称：%v", err, dmt.FileName))
			// 触发下载异常
			dlErrorHandler(dmt)
			return
		}

		// 其他下载异常
		mylog.Error(fmt.Sprintf("下载失败：%v，重新加入下载列表", err))
		offerBack(dmt)
	})
}

// initCoreDownloader 根据全局配置初始化一个下载器对象
// 如果初始化失败，程序将直接退出
func initCoreDownloader() coredl.Downloader {

	// 获取配置
	resource := config.G.Decoder.ResourceType
	dlType := config.G.Downloader.Use

	// 生成对象
	switch resource + dlType {

	case config.ResourceMP4 + config.DownloadSimple:
		return coredl.NewMp4Simple()

	case config.ResourceMP4 + config.DownloadMultiThread:
		return coredl.NewMp4MultiThread()

	case config.ResourceM3U8 + config.DownloadSimple:
		return coredl.NewM3U8Simple()

	case config.ResourceM3U8 + config.DownloadMultiThread:
		return coredl.NewM3U8MultiThread()

	default:
		log.Fatal("下载器初始化异常，请检查配置")
		return nil
	}
}
