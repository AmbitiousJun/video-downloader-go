// 转换器配置
package config

import (
	"fmt"
	"strings"
	"syscall"
	"video-downloader-go/internal/util/mylog/color"

	"github.com/pkg/errors"
)

type Transfer struct {
	Use             string `yaml:"use"`               // 要选用哪个转换器，可选值：ffmpeg
	TsFilenameRegex string `yaml:"ts-filename-regex"` // 正则表达式，用于匹配出 ts 文件的序号
}

const (
	TransferFfmpegStr   = "ffmpeg_str"    // ffmpeg 转换器
	TransferFfmpegStrV2 = "ffmpeg_str_v2" // ffmpeg 转换器
	TransferFfmpegTxt   = "ffmpeg_txt"    // ffmpeg 转换器
)

// 默认的 ts 文件名序号匹配正则
const DefaultFilenameRegex = "_(\\d+)\\."

// checkFields 检查转换器字段是否合法
func (t *Transfer) checkFields(allowEmpty bool) error {
	t.Use = strings.TrimSpace(t.Use)
	validTypes := []string{TransferFfmpegStr, TransferFfmpegTxt, TransferFfmpegStrV2}

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
	increaseSystemUlimit(65535)
	return nil
}

// increaseSystemUlimit 增大系统最多可打开的文件描述符个数
func increaseSystemUlimit(limit uint64) {
	var rLimit syscall.Rlimit

	// 获取当前文件描述符限制
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Printf(color.ToRed("修改文件描述符最大个数失败: %v"), err)
		return
	}

	// 更新文件描述符限制
	rLimit.Cur = limit
	if limit > rLimit.Max {
		rLimit.Max = limit
	}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Printf(color.ToRed("修改文件描述符最大个数失败: %v"), err)
	}
}
