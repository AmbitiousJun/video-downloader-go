package transfer

import (
	"sync"
	"video-downloader-go/src/config"
)

// ts 文件转换器接口
type TsTransfer interface {
	// 将 ts 格式的文件列表转换成 mp4 格式的视频文件
	// @param tsDir 存放 ts 文件的目录
	// @param outputPath 合并后输出的文件绝对地址
	Ts2Mp4(tsDir, outputPath string) error
}

func NewFfmpegTransfer() TsTransfer {
	return &ffmpegTransfer{}
}

// 获取 ts 转换器
var Instance = (func() func() TsTransfer {
	var tsTransfer TsTransfer = nil
	var mu sync.Mutex
	return func() TsTransfer {
		if tsTransfer == nil {
			mu.Lock()
			defer mu.Unlock()
			if tsTransfer == nil {
				tfType := config.GlobalConfig.Transfer.Use
				switch tfType {
				case config.TransferFfmpeg:
					tsTransfer = NewFfmpegTransfer()
				default:
					panic("没有初始化 ts 转换器类型")
				}
			}
		}
		return tsTransfer
	}
})()
