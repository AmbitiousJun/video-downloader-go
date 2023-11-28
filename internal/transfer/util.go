package transfer

import (
	"sync"
	"video-downloader-go/internal/config"
)

// 获取一个全局唯一的 ts 转换器
var Instance = (func() func() TsTransfer {
	var t TsTransfer = nil
	var once sync.Once
	return func() TsTransfer {
		once.Do(func() {
			tfType := config.GlobalConfig.Transfer.Use
			switch tfType {
			case config.TransferFfmpeg:
				t = &ffmpegTransfer{}
			default:
				panic("没有初始化 ts 转换器类型")
			}
		})
		return t
	}
})()
