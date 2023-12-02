package main

import (
	"fmt"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"
)

func main() {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	err := config.Load("")
	if err != nil {
		mylog.Error(err.Error())
		return
	}
	fmt.Println(config.GlobalConfig)
}
