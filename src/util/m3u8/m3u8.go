package m3u8

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"video-downloader-go/src/util/log"
)

const (
	// http 地址前缀
	NetworkLinkPrefix = "http"
	// 本地文件地址前缀
	LocalFilePrefix = "file"
)

// 响应头中，有效的 m3u8 Content-Length 属性
var ValidM3U8ContentTypes = map[string]struct{}{
	"application/vnd.apple.mpegurl": {},
	"application/x-mpegurl":         {},
}

// 检查一个 url 是否是 m3u8 地址
// @url: 要检查的地址
// @headers: 附加的请求头
func CheckM3U8(url string, headers map[string]string) bool {
	if len(url) == 0 {
		return false
	}
	for {
		log.Info("正在解析 m3u8 信息...")
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Warn(fmt.Sprintf("解析异常：%v，两秒后重试", err.Error()))
			time.Sleep(2000)
			continue
		}
		// 添加请求头
		for k, v := range headers {
			request.Header.Set(k, v)
		}
		request.Header.Set("Connection", "Close")
		// 创建客户端，发送请求
		client := http.DefaultClient
		resp, err := client.Do(request)
		if err != nil {
			log.Warn(fmt.Sprintf("解析异常：%v，两秒后重试", err.Error()))
			time.Sleep(2000)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Warn(fmt.Sprintf("解析异常：错误码 %v，两秒后重试", resp.StatusCode))
			time.Sleep(2000)
			continue
		}
		contentType := resp.Header.Get("Content-Type")
		if len(contentType) == 0 {
			log.Warn(fmt.Sprintf("解析异常：%v，两秒后重试", "无法获取目标的 Content-Type 属性"))
			time.Sleep(2000)
			continue
		}
		contentType = strings.Split(strings.ToLower(contentType), ";")[0]
		_, valid := ValidM3U8ContentTypes[contentType]
		return valid
	}
}
