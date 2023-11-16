// 转换器配置
package config

import (
	"strings"

	"github.com/pkg/errors"
)

type Transfer struct {
	Use             string `yaml:"use"`               // 要选用哪个转换器，可选值：ffmpeg
	TsFilenameRegex string `yaml:"ts-filename-regex"` // 正则表达式，用于匹配出 ts 文件的序号
}

const (
	TransferFfmpeg = "ffmpeg" // ffmpeg 转换器
)

// 默认的 ts 文件名序号匹配正则
const DefaultFilenameRegex = "(?<=_)(\\d+)(?=\\.)"

// 检查转换器配置
func checkTransferConfig() error {
	cfg := GlobalConfig.Transfer
	cfg.Use = strings.TrimSpace(cfg.Use)
	validTypes := []string{TransferFfmpeg}
	flag := false
	for _, valid := range validTypes {
		if valid == cfg.Use {
			flag = true
		}
	}
	if !flag {
		return errors.New("转换器类型配置错误，可选值：ffmpeg")
	}
	cfg.TsFilenameRegex = strings.TrimSpace(cfg.TsFilenameRegex)
	if cfg.TsFilenameRegex == "" {
		cfg.TsFilenameRegex = DefaultFilenameRegex
	}
	if cfg.Use == TransferFfmpeg {
		err := checkFfmpegEnv()
		return errors.Wrap(err, "转换器配置错误")
	}
	return nil
}

// 检查 ffmpeg 环境
func checkFfmpegEnv() error {
	// TODO: 检查 ffmpeg 可执行文件是否能够正常使用
	return nil
}
