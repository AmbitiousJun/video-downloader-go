package main

import (
	"fmt"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/log"
)

func main() {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	err := config.Load("")
	if err != nil {
		log.Error(err.Error())
		return
	}
	fmt.Println(config.GlobalConfig)
}
