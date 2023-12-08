// 使用命令行的方式执行下载器

package downloader

import (
	"fmt"
	"path/filepath"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
)

// 错误信息
const (
	UnSupportDownloadType = "不支持的下载方式"
	UnValidM3U8           = "不是规范的 m3u8 文件"
)

// 任务下载完成的监听器
type DownloadListener func()

// 任务下载失败的监听器，下载器会将失败的任务传递出来
type DownloadErrorListener func(dmt *meta.Download)

// ListenAndDownload 用于命令行模式下监听下载任务并依据全局配置多协程下载任务
func ListenAndDownload(list *meta.TaskDeque[meta.Download], listener DownloadListener, errListener DownloadErrorListener) {
	appctx.WaitGroup().Add(1)
	go func() {
		defer appctx.WaitGroup().Done()
		select {
		case <-appctx.Context().Done():
			return
		default:
			doListen(list, listener, errListener)
		}
	}()
}

// doListen 是监听下载任务的核心函数，它运行在一个独立的 goroutine 上
func doListen(list *meta.TaskDeque[meta.Download], completeOne DownloadListener, emitError DownloadErrorListener) {
	for {
		mylog.Info("开始监听下载列表...")
		for list.Empty() {
			// 每隔两秒检查一下是否有新的下载任务
			time.Sleep(time.Second * 2)
		}
		dmt := list.PollFirst()
		// TODO：增加协程池判空校验
		dlpool.Task.Submit(func() {
			appctx.WaitGroup().Add(1)
			defer appctx.WaitGroup().Done()
			originFilename := dmt.FileName
			link := dmt.Link
			fileName := config.GlobalConfig.Downloader.DownloadDir + string(filepath.Separator) + originFilename + ".mp4"
			dmt.FileName = fileName
			mylog.Info(fmt.Sprintf("监听到下载任务，文件名：%v，下载地址：%v", fileName, link))
			// TODO：初始化下载器并下载
			completeOne()
		})
	}
}
