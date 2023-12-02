package mylog_test

import (
	"testing"
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
