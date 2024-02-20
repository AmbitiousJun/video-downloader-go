package mylog_test

import (
	"sync"
	"testing"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/util/mylog"
)

func TestLog(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	mylog.Info("测试 Info 日志")
	mylog.Error("测试 Error 日志")
	mylog.Warn("测试 Warn 日志")
	mylog.Success("测试 Success 日志")
}

func TestDownloadLog(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()

	var wg sync.WaitGroup
	defer wg.Wait()

	var d *mylog.DownloadLog
	for i := 1; i <= 3; i++ {
		dl := mylog.NewDownloadLog()
		dl.Progress("文件名: .....")
		dl.Progress("下载进度: ....")
		dl.Progress("进度条: ...")
		dl.Trigger()
		d = dl
	}

	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		currentI := i
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * time.Duration(currentI))
			d.Trigger()
		}()
	}
}
