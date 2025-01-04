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
	TransferFfmpegStr = "ffmpeg_str" // ffmpeg 转换器
	TransferFfmpegTxt = "ffmpeg_txt" // ffmpeg 转换器
)

// 默认的 ts 文件名序号匹配正则
const DefaultFilenameRegex = "_(\\d+)\\."

// checkFields 检查转换器字段是否合法
func (t *Transfer) checkFields(allowEmpty bool) error {
	t.Use = strings.TrimSpace(t.Use)
	validTypes := []string{TransferFfmpegStr, TransferFfmpegTxt}

	if t.Use == "" && !allowEmpty {
		return errors.New("转换器类型配置错误，可选值：" + strings.Join(validTypes, ","))
	}

	flag := false
	if t.Use != "" {
		for _, valid := range validTypes {
			if valid == t.Use {
				flag = true
				break
			}
		}
		if !flag {
			return errors.New("转换器类型配置错误，可选值：" + strings.Join(validTypes, ","))
		}
	}

	t.TsFilenameRegex = strings.TrimSpace(t.TsFilenameRegex)
	if t.TsFilenameRegex == "" {
		t.TsFilenameRegex = DefaultFilenameRegex
	}
	return nil
}

// 检查转换器配置
func checkTransferConfig() error {
	cfg := G.Transfer
	if err := cfg.checkFields(false); err != nil {
		return err
	}
	return nil
}
