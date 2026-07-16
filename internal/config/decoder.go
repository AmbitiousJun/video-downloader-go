// 解析器配置
package config

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"video-downloader-go/internal/util/mylog"

	"github.com/pkg/errors"
)

const (
	DecoderNone       = "none"
	DecoderYoutubeDl  = "youtube-dl"
	DecoderCatCatchTx = "cat-catch:tx"
	DecoderCatCatchMg = "cat-catch:mg"
)

const (
	ResourceMP4  = "mp4"
	ResourceM3U8 = "m3u8"
)

const (
	YoutubeDlCookieNone = "none" // 配置为 none 表示不注入 cookie

	YoutubeDlRememberFormatActive   = 1  // 记住解析记录 激活
	YoutubeDlRememberFormatDeactive = -1 // 记住解析记录 不激活

	CatCatchHeadlessActive   = 1  // 猫抓解析器开启无头模式
	CatCatchHeadlessDeactive = -1 // 猫抓解析器关闭无头模式
)

type Decoder struct {
	Use       string          `yaml:"use"`        // 使用哪种解析方式，可选值：none, youtube-dl, cat-catch:tx, cat-catch:mg
	MaxRetry  int             `yaml:"max-retry"`  // 最大的尝试解析次数
	YoutubeDL YoutubeDlConfig `yaml:"youtube-dl"` // youtube-dl 解析器相关配置
	CatCatch  CatCatchConfig  `yaml:"cat-catch"`  // cat-catch 解析器
}

type YoutubeDlConfig struct {
	CookiesFrom    string   `yaml:"cookies-from"`    // 从哪个浏览器获取 cookie
	RawFormatCodes []string `yaml:"format-codes"`    // 下载视频的编码
	RememberFormat int      `yaml:"remember-format"` // 是否记住视频格式
	FormatCodes    []*YtDlFormatCode
}

// 经过校验并封装的 youtube-dl format code
type YtDlFormatCode struct {
	Code             string
	ExpectedLinkNums int
}

// CatCatchSiteConfig 猫抓解析器网站配置
type CatCatchSiteConfig struct {
	CookieJsonPath string `yaml:"cookie-json-path"` // Cookie 文件绝对路径
	VideoFormat    string `yaml:"video-format"`     // 视频格式
}

// 猫抓解析器配置
type CatCatchConfig struct {
	Headless int `yaml:"headless"` // 是否开启无头模式
	Sites    struct {
		Tx CatCatchSiteConfig `yaml:"tx"`
		Mg CatCatchSiteConfig `yaml:"mg"`
	} `yaml:"sites"` // 猫抓解析器需要对每个站点单独适配
}

// 检查解析器配置
func checkDecoderConfig() error {
	cfg := &G.Decoder

	if err := cfg.checkFields(false); err != nil {
		return errors.Wrap(err, "解析器配置异常")
	}

	// 检查 youtube-dl 环境
	if err := checkYtDlEnv(); err != nil {
		return errors.Wrap(err, "检查 youtube-dl 环境失败")
	}

	return nil
}

// checkFields 检查 Decoder 对象中的属性值是否合法
// 不检查系统环境
// allowEmpty 参数为 true 时，对于空值不进行校验，也不返回错误
func (dc *Decoder) checkFields(allowEmpty bool) error {
	// 1 检查解析器类型是否合法
	validTypes := []string{DecoderNone, DecoderYoutubeDl, DecoderCatCatchTx, DecoderCatCatchMg}
	dc.Use = strings.TrimSpace(dc.Use)

	if dc.Use == "" && !allowEmpty {
		return errors.New("解析器类型配置错误，可选值：" + strings.Join(validTypes, ","))
	}

	flag := false

	if dc.Use != "" {
		if slices.Contains(validTypes, dc.Use) {
			flag = true
		}
		if !flag {
			return errors.New("解析器类型配置错误，可选值：" + strings.Join(validTypes, ","))
		}
	}

	// 3 检查 format code
	if err := dc.checkFormatCodes(); err != nil {
		return errors.Wrap(err, "检查 format code 失败")
	}

	// 4 设置默认的 cookie 来源
	dc.YoutubeDL.CookiesFrom = strings.TrimSpace(dc.YoutubeDL.CookiesFrom)
	if dc.YoutubeDL.CookiesFrom == "" {
		dc.YoutubeDL.CookiesFrom = YoutubeDlCookieNone
	}

	// 5 检查 youtube-dl 记住视频格式配置
	if !dc.YoutubeDL.IsRememberFormatValid() && !allowEmpty {
		return errors.New("remember format 配置错误，可选值: -1, 1")
	}

	// 6 检查猫抓解析器配置
	if !dc.CatCatch.IsHeadlessValid() && !allowEmpty {
		return errors.New("headless 配置错误, 可选择: -1, 1")
	}

	// 7 检查最大重试次数
	if dc.MaxRetry < 1 && !allowEmpty {
		return errors.New("max-retry 配置错误, 必须大于 1")
	}

	return nil
}

// IsHeadlessValid 检查用户配置的 headless 配置是否有效
func (c *CatCatchConfig) IsHeadlessValid() bool {
	valids := []int{CatCatchHeadlessActive, CatCatchHeadlessDeactive}
	return slices.Contains(valids, c.Headless)
}

// IsRememberFormatValid 检查对象中的 RememberFormat 属性是否有效
func (c *YoutubeDlConfig) IsRememberFormatValid() bool {
	validRfs := []int{YoutubeDlRememberFormatDeactive, YoutubeDlRememberFormatActive}
	return slices.Contains(validRfs, c.RememberFormat)
}

// 检查 format code
func (dc *Decoder) checkFormatCodes() error {
	rawCodes := dc.YoutubeDL.RawFormatCodes

	formatCodes := []*YtDlFormatCode{}
	for _, raw := range rawCodes {
		cs := strings.Split(raw, "+")
		if len(cs) != 1 && len(cs) != 2 {
			return errors.New(fmt.Sprintf("不合法的 format code：%v，示例：137+140", raw))
		}
		formatCodes = append(formatCodes, &YtDlFormatCode{raw, len(cs)})
	}
	dc.YoutubeDL.FormatCodes = formatCodes
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
	mylog.Success("检查 youtube-dl 环境成功")
	return nil
}
