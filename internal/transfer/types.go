package transfer

import "video-downloader-go/internal/util/mylog/dlbar"

// ts 文件转换器接口
type TsTransfer interface {
	// 将 ts 格式的文件列表转换成 mp4 格式的视频文件
	// @param tsDir 存放 ts 文件的目录
	// @param outputPath 合并后输出的文件绝对地址
	// @param bar 任务日志
	Ts2Mp4(tsDir, outputPath string, bar *dlbar.Bar) error
}
