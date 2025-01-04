// 不同域名的定制化配置

package config

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type CustomConfig struct {
	Decoder  Decoder  `yaml:"decoder"`  // 解析器配置
	Transfer Transfer `yaml:"transfer"` // 转换器配置
	Hosts    []string `yaml:"hosts"`    // 指定的域名列表
}

// host2Decoder 保存每个 host 的定制化解析器配置
var host2Decoder map[string]*Decoder

// host2Transfer 保存每个 host 的定制化转换器配置
var host2Transfer map[string]*Transfer

// checkCustomConfig 执行定制化配置的初始化
func checkCustomConfig() error {
	host2Decoder = make(map[string]*Decoder)
	host2Transfer = make(map[string]*Transfer)
	customs := G.Customs

	for i, custom := range customs {
		copyCustom := custom

		// 1 检查解析器配置
		if err := copyCustom.Decoder.checkFields(true); err != nil {
			return errors.Wrapf(err, "请检查定制化的解析器配置, index: %v", i)
		}

		// 2 检查转换器配置
		if err := copyCustom.Transfer.checkFields(true); err != nil {
			return errors.Wrapf(err, "请检查定制化的解析器配置, index: %v", i)
		}

		// 3 保存 host 映射
		for _, host := range copyCustom.Hosts {
			h := strings.TrimSpace(host)
			if h == "" {
				continue
			}
			host2Decoder[h] = &copyCustom.Decoder
			host2Transfer[h] = &copyCustom.Transfer
		}
	}

	return nil
}

// CustomUse 返回一个解析器类型
// 优先返回定制化配置
func (dc *Decoder) CustomUse(dcUrl string) string {
	targetDecoder := resolveDecoderByUrl(dcUrl, dc)
	if targetDecoder == nil || targetDecoder.Use == "" {
		return dc.Use
	}

	return targetDecoder.Use
}

// CustomMaxRetry 返回解析器最大的重试次数
// 优先返回定制化配置
func (dc *Decoder) CustomMaxRetry(dcUrl string) int {
	targetDecoder := resolveDecoderByUrl(dcUrl, dc)
	if targetDecoder == nil || targetDecoder.MaxRetry < 1 {
		return dc.MaxRetry
	}

	return targetDecoder.MaxRetry
}

// CustomCookiesFrom 返回一个 youtube-dl 的 cookie 来源
// 优先返回定制化配置
func (y *YoutubeDlConfig) CustomCookiesFrom(dcUrl string) string {
	targetDecoder := resolveDecoderByUrl(dcUrl, nil)
	if targetDecoder == nil || targetDecoder.YoutubeDL.CookiesFrom == "" {
		return y.CookiesFrom
	}

	return targetDecoder.YoutubeDL.CookiesFrom
}

// CustomFormatCode 返回 youtube-dl 预配置的 format code 列表
// 优先返回定制化配置
func (y *YoutubeDlConfig) CustomFormatCodes(dcUrl string) []*YtDlFormatCode {
	targetDecoder := resolveDecoderByUrl(dcUrl, nil)
	if targetDecoder == nil || len(targetDecoder.YoutubeDL.FormatCodes) == 0 {
		return y.FormatCodes
	}

	return targetDecoder.YoutubeDL.FormatCodes
}

// CustomRememberFormat 返回使用 youtube-dl 解析器是否自动记住上一次解析结果
func (y *YoutubeDlConfig) CustomRememberFormat(dcUrl string) int {
	targetDecoder := resolveDecoderByUrl(dcUrl, nil)
	if targetDecoder == nil || !targetDecoder.YoutubeDL.IsRememberFormatValid() {
		return y.RememberFormat
	}

	return targetDecoder.YoutubeDL.RememberFormat
}

// CustomHeadless 返回使用 chromedriver 时是否开启无头模式
func (c *CatCatchConfig) CustomHeadless(dcUrl string) int {
	targetDecoder := resolveDecoderByUrl(dcUrl, nil)
	if targetDecoder == nil || !targetDecoder.CatCatch.IsHeadlessValid() {
		return c.Headless
	}
	return targetDecoder.CatCatch.Headless
}

// resolveDecoderByUrl 根据解析 url 返回解析器
// 优先返回定制化配置解析器
func resolveDecoderByUrl(dcUrl string, defaultDecoder *Decoder) *Decoder {
	if defaultDecoder == nil {
		// 默认使用全局解析器
		defaultDecoder = &G.Decoder
	}

	u, err := url.Parse(dcUrl)
	if err != nil {
		// 解析 url 异常
		return defaultDecoder
	}

	if target, ok := host2Decoder[u.Host]; ok {
		// 成功找到匹配的定制化解析器配置
		return target
	}

	return defaultDecoder
}

// CustomUse 优先使用定制化的转换器类型
func (t *Transfer) CustomUse(originUrl string) string {
	targetTransfer := resolveTransferByUrl(originUrl, nil)
	if targetTransfer == nil || targetTransfer.Use == "" {
		return t.Use
	}
	return targetTransfer.Use
}

// resolveTransferByUrl 根据解析 url 返回解析器
// 优先返回定制化配置解析器
func resolveTransferByUrl(originUrl string, defaultTransfer *Transfer) *Transfer {
	if defaultTransfer == nil {
		// 默认使用全局转换器
		defaultTransfer = &G.Transfer
	}

	u, err := url.Parse(originUrl)
	if err != nil {
		// 解析 url 异常
		return defaultTransfer
	}

	if target, ok := host2Transfer[u.Host]; ok {
		// 成功找到匹配的定制化解析器配置
		return target
	}

	return defaultTransfer
}
