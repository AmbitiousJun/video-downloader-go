package log_test

import (
	"context"
	"sync"
	"testing"
	"video-downloader-go/src/util/log"
)

func TestLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	log.InitLog(ctx, &wg)
	log.Info("测试 Info 日志")
	log.Error("测试 Error 日志")
	log.Warn("测试 Warn 日志")
	log.Success("测试 Success 日志")
	cancel()
	wg.Wait()
}
