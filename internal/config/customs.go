// 不同域名的定制化配置

package config

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type CustomConfig struct {
	Decoder Decoder  `yaml:"decoder"` // 解析器配置
	Hosts   []string `yaml:"hosts"`   // 指定的域名列表
}

// host2Decoder 保存每个 host 的定制化解析器配置
var host2Decoder map[string]*Decoder

// checkCustomConfig 执行定制化配置的初始化
func checkCustomConfig() error {
	host2Decoder = make(map[string]*Decoder)
	customs := G.Customs

	for i, custom := range customs {
		copyCustom := custom

		// 1 检查解析器配置
		if err := copyCustom.Decoder.checkFields(true); err != nil {
			return errors.Wrapf(err, "请检查定制化的解析器配置, index: %v", i)
		}

		// 2 保存 host 映射
		for _, host := range copyCustom.Hosts {
			h := strings.TrimSpace(host)
			if h == "" {
				continue
			}
			host2Decoder[h] = &copyCustom.Decoder
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

// CustomResourceType 返回一个媒体类型
// 优先返回定制化配置
func (dc *Decoder) CustomResourceType(dcUrl string) string {
	targetDecoder := resolveDecoderByUrl(dcUrl, dc)
	if targetDecoder == nil || targetDecoder.ResourceType == "" {
		return dc.ResourceType
	}

	return targetDecoder.ResourceType
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
