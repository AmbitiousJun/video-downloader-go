// 解析器配置
package config

import (
	"fmt"
	"os/exec"
	"strings"
	"video-downloader-go/src/util/log"

	"github.com/pkg/errors"
)

const (
	DecoderNone      = "none"
	DecoderYoutubeDl = "youtube-dl"
)

const (
	ResourceMP4  = "mp4"
	ResourceM3U8 = "m3u8"
)

// 配置为 none 表示不注入 cookie
const YoutubeDlCookieNone = "none"

type Decoder struct {
	Use          string          `yaml:"use"`           // 使用哪种解析方式，可选值：none, youtube-dl
	ResourceType string          `yaml:"resource-type"` // 解析出来的地址类型，可选值：mp4, m3u8
	YoutubeDL    YoutubeDlConfig `yaml:"youtube-dl"`    // youtube-dl 解析器相关配置
}

type YoutubeDlConfig struct {
	CookiesFrom    string   `yaml:"cookies-from"` // 从哪个浏览器获取 cookie
	RawFormatCodes []string `yaml:"format-codes"` // 下载视频的编码
	FormatCodes    []*YtDlFormatCode
}

// 经过校验并封装的 youtube-dl format code
type YtDlFormatCode struct {
	Code             string
	ExpectedLinkNums int
}

// 检查解析器配置
func checkDecoderConfig() error {
	cfg := GlobalConfig.Decoder
	// 1 检查解析器类型是否合法
	validTypes := []string{DecoderNone, DecoderYoutubeDl}
	cfg.Use = strings.TrimSpace(cfg.Use)
	flag := false
	for _, valid := range validTypes {
		if strings.EqualFold(valid, cfg.Use) {
			flag = true
		}
	}
	if !flag {
		return errors.New("解析器类型配置错误，可选值：none, youtube-dl")
	}
	// 2 检查资源类型是否合法
	validResources := []string{ResourceMP4, ResourceM3U8}
	cfg.ResourceType = strings.TrimSpace(cfg.ResourceType)
	flag = false
	for _, valid := range validResources {
		if strings.EqualFold(valid, cfg.ResourceType) {
			flag = true
		}
	}
	if !flag {
		return errors.New("媒体类型配置错误，可选值：mp4，m3u8")
	}
	// 3 检查 youtube-dl 环境
	var err error
	if strings.EqualFold(cfg.Use, DecoderYoutubeDl) {
		err = checkYtDlEnv()
		if err != nil {
			return errors.Wrap(err, "检查 youtube-dl 环境失败")
		}
	}
	// 4 检查 format code
	err = checkFormatCodes()
	if err != nil {
		return errors.Wrap(err, "检查 format code 失败")
	}
	// 5 设置默认的 cookie 来源
	cfg.YoutubeDL.CookiesFrom = strings.TrimSpace(cfg.YoutubeDL.CookiesFrom)
	if cfg.YoutubeDL.CookiesFrom == "" {
		cfg.YoutubeDL.CookiesFrom = YoutubeDlCookieNone
	}
	return nil
}

// 检查 format code
func checkFormatCodes() error {
	rawCodes := GlobalConfig.Decoder.YoutubeDL.RawFormatCodes
	formatCodes := []*YtDlFormatCode{}
	for _, raw := range rawCodes {
		cs := strings.Split(raw, "+")
		if len(cs) != 1 && len(cs) != 2 {
			return errors.New(fmt.Sprintf("不合法的 format code：%v，示例：137+140", raw))
		}
		formatCodes = append(formatCodes, &YtDlFormatCode{raw, len(cs)})
	}
	GlobalConfig.Decoder.YoutubeDL.FormatCodes = formatCodes
	return nil
}

// 检查 youtube-dl 环境
func checkYtDlEnv() error {
	cmd := exec.Command(YoutubeDlPath, "--help")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	result := string(output)
	if !strings.Contains(result, "Usage:") {
		return errors.New("无法执行命令")
	}
	log.Success("检查 youtube-dl 环境成功")
	return nil
}
