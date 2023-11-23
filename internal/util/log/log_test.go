package log_test

import (
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/util/log"
)

func TestLog(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	log.Info("测试 Info 日志")
	log.Error("测试 Error 日志")
	log.Warn("测试 Warn 日志")
	log.Success("测试 Success 日志")
}
