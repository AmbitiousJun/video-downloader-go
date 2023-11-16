// 解析器配置
package config

type Decoder struct {
	Use          string          `yaml:"use"`           // 使用哪种解析方式，可选值：none, youtube-dl
	ResourceType string          `yaml:"resource-type"` // 解析出来的地址类型，可选值：mp4, m3u8
	YoutubeDL    YoutubeDlConfig `yaml:"youtube-dl"`    // youtube-dl 解析器相关配置
}

type YoutubeDlConfig struct {
	CookiesFrom string   `yaml:"cookies-from"` // 从哪个浏览器获取 cookie
	FormatCodes []string `yaml:"format-codes"` // 下载视频的编码
}

const (
	DecoderNone      = "none"
	DecoderYoutubeDl = "youtube-dl"
)

const (
	ResourceMP4  = "mp4"
	ResourceM3U8 = "m3u8"
)

// 检查解析器配置
func checkDecoderConfig() error {
	return nil
}
