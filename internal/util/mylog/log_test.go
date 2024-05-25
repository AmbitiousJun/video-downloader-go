package mylog_test

import (
	"fmt"
	"testing"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mylog/dlbar"
)

func TestPanel(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	b := dlbar.NewBar(
		dlbar.WithPercent(0),
		dlbar.WithSize(0),
		dlbar.WithStatus(dlbar.BarStatusExecuting),
		dlbar.WithChildStatus(dlbar.BarChildStatusDownload),
		dlbar.WithName("The.Truth.S02E08.2024-05-24.第4期下：《走不出的忙活街》-抓马认亲"),
	)
	fmt.Println("这是一条示例日志")
	fmt.Println("这是一条示例日志")
	fmt.Println("这是一条示例日志")
	fmt.Println("这是一条示例日志")
	fmt.Println("这是一条示例日志")
	mylog.GlobalPanel.RegisterBar(b)
	mylog.Start()
	for i := 1; i <= 4; i++ {
		time.Sleep(5 * time.Second)
		b.UpdatePercentAndSize(i*25, int64(i*536870912))
	}
	b.OkHint("下载完成")
}
