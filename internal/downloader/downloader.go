// 使用命令行的方式执行下载器

package downloader

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/downloader/ytdl"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/m3u8"
	"video-downloader-go/internal/util/myfile"
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

// CanDownloadChan 当下载器成功下载一个任务时, 就往该通道中写入信号, 通知解析器及时解析
var CanDownloadChan = make(chan struct{}, 1)

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
		dmt.LogBar.UpdatePercentAndSize(0, 0)

		// 初始化下载器并下载
		cdl := initCoreDownloader(dmt)
		progressMap := make(map[int]*coredl.Progress)
		progressMu := sync.Mutex{}
		err := cdl.Exec(dmt, func(p *coredl.Progress) {
			progressMu.Lock()
			defer progressMu.Unlock()
			progressMap[p.CurrentTask] = p
			percent, size := 0, int64(0)
			current, total := float64(0), float64(0)
			for _, progress := range progressMap {
				// 1 size 累加
				size += progress.CurrentBytes
				// 2 字节数不一致
				if progress.CurrentBytes != progress.TotalBytes {
					current += float64(progress.CurrentBytes)
					total += float64(progress.TotalBytes)
					continue
				}
				// 3 字节数一致, 按照分片数统计
				current += float64(progress.Current)
				total += float64(progress.Total)
			}
			percent = int(math.Round((current / total) * float64(100)))
			dmt.LogBar.UpdatePercentAndSize(percent, size)
		})

		// 下载成功
		if err == nil {
			completeOne()
			dmt.LogBar.OkHint("下载完成")
			select {
			case CanDownloadChan <- struct{}{}:
			default:
			}
			return
		}

		// 下载出现异常，检查是否有下载一半的文件，将其删除
		myfile.DeleteAnyFileContainsPrefix(dmt.FileName)
		// 恢复原始的下载文件名
		dmt.FileName = originFilename

		// 下载失败，无效的 m3u8
		if strings.Contains(err.Error(), UnValidM3U8) {
			mylog.Warnf("下载失败：%v, 重新添加到解析任务中，视频名称：%v", err, dmt.FileName)
			dmt.LogBar.ErrorHint("下载失败, 等待重新解析")
			// 触发下载异常
			dlErrorHandler(dmt)
			return
		}

		// 其他下载异常
		mylog.Errorf("下载失败：%v，重新加入下载列表", err)
		dmt.LogBar.ErrorHint("下载失败, 等待重新下载")
		offerBack(dmt)
	})
}

// initCoreDownloader 根据全局配置初始化下载器对象
// 优先匹配定制化配置
func initCoreDownloader(dmt *meta.Download) coredl.Downloader {

	// 如果是通过 youtube-dl 解析的，就使用适配的下载器
	if config.G.Decoder.CustomUse(dmt.OriginUrl) == config.DecoderYoutubeDl {
		return ytdl.New()
	}

	// 获取配置
	dlType := config.G.Downloader.Use

	// 识别资源类型
	resource := config.ResourceMP4
	if m3u8.CheckM3U8(dmt.Link, dmt.HeaderMap) {
		resource = config.ResourceM3U8
	}

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
	}

	return nil
}
