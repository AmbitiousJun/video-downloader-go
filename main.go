package main

import (
	"fmt"
	"video-downloader-go/src/appctx"
	"video-downloader-go/src/config"
	"video-downloader-go/src/util/log"
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
