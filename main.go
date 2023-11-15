package main

import (
	"fmt"
	"video-downloader-go/src/appctx"
	"video-downloader-go/src/config"
	"video-downloader-go/src/util/log"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Error(err.Error())
	}
	fmt.Println(config.GlobalConfig)
	appctx.CancelFunc()()
	appctx.WaitGroup().Wait()
}
