package m3u8

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"video-downloader-go/src/entity"
	"video-downloader-go/src/util/log"

	"github.com/pkg/errors"
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

// 读取 M3U8 文件中的 ts 文件列表
// @m3u8url: m3u8 文件的下载地址
// @headers 请求头，可以为空
// @return ts 文件列表
func ReadTsUrls(m3u8Url string, headers map[string]string) ([]*entity.TsMeta, error) {
	if strings.HasPrefix(m3u8Url, NetworkLinkPrefix) {
		return readHttpTsUrls(m3u8Url, headers)
	}
	stat, err := os.Stat(m3u8Url[7:])
	for err != nil || stat.IsDir() {
		// 如果有错误，说明文件不存在，重复判断
		log.Info("查找不到本地的 m3u8 文件：" + m3u8Url)
		time.Sleep(1000)
		stat, err = os.Stat(m3u8Url[7:])
	}
	// 1 读取数据
	mUrl, err := url.Parse(m3u8Url)
	if err != nil {
		return nil, errors.Wrap(err, "转换本地 m3u8 url 出现异常")
	}
	mFile, err := os.Open(mUrl.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "打开本地 m3u8 文件出现异常，path: %v", mUrl.Path)
	}
	defer mFile.Close()
	ans := []*entity.TsMeta{}
	scanner := bufio.NewScanner(mFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			// 去除掉注释和空行
			continue
		}
		if !strings.HasPrefix(line, NetworkLinkPrefix) {
			return nil, errors.New("m3u8 文件不规范：检测不到 http 协议")
		}
		// 2 封装对象
		ans = append(ans, entity.NewTsMeta(line, len(ans)+1))
	}
	// 3 TODO 删除 m3u8 文件
	return ans, nil
}

// 读取网络 M3U8 文件
// @m3u8Url: url
// @headers: 请求头
// @return ts urls
func readHttpTsUrls(m3u8Url string, headers map[string]string) ([]*entity.TsMeta, error) {
	return nil, nil
}
