package transfer

import (
	"video-downloader-go/internal/config"
)

// 获取 ts 转换器实例
// 根据 originUrl 使用不同的转换器
func Instance(originUrl string) TsTransfer {
	tfType := config.G.Transfer.CustomUse(originUrl)
	switch tfType {
	case config.TransferFfmpegStr:
		return &ffmpegTransfer{concatFileFunc: ConcatFilesByStr}
	case config.TransferFfmpegTxt:
		return &ffmpegTransfer{concatFileFunc: ConcatFilesByTxt}
	default:
		panic("没有初始化 ts 转换器类型")
	}
}
