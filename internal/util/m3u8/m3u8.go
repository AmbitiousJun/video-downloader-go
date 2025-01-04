package m3u8

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/transfer"
	"video-downloader-go/internal/util"
	"video-downloader-go/internal/util/myhttp"
	"video-downloader-go/internal/util/mylog"

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
// @param url 要检查的地址
// @param headers 附加的请求头
// @return 是否是一个有效的 m3u8 地址
func CheckM3U8(url string, headers map[string]string) bool {
	if len(url) == 0 {
		return false
	}
	for {
		mylog.Info("正在解析 m3u8 信息...")
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			util.PrintRetryError("构造请求失败", err, 2)
			continue
		}
		// 添加请求头
		for k, v := range headers {
			request.Header.Set(k, v)
		}
		request.Header.Set("Connection", "Close")
		// 创建客户端，发送请求
		client := myhttp.TimeoutHttpClient()
		resp, err := client.Do(request)
		if err != nil {
			util.PrintRetryError("发送请求异常", err, 2)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			util.PrintRetryError("请求响应异常", err, 2)
			continue
		}
		contentType := resp.Header.Get("Content-Type")
		if len(contentType) == 0 {
			util.PrintRetryError("解析异常", errors.New("无法获取目标的 Content-Type 属性"), 2)
			continue
		}
		contentType = strings.Split(strings.ToLower(contentType), ";")[0]
		_, valid := ValidM3U8ContentTypes[contentType]
		return valid
	}
}

// 读取 M3U8 文件中的 ts 文件列表
// @param m3u8url m3u8 文件的下载地址
// @param headers 请求头，可以为空
// @return ts 文件列表
func ReadTsUrls(m3u8Url string, headers map[string]string) ([]*TsMeta, error) {
	if strings.HasPrefix(m3u8Url, NetworkLinkPrefix) {
		return readHttpTsUrls(m3u8Url, headers)
	}
	prefix := LocalFilePrefix + "://"
	if !strings.HasPrefix(m3u8Url, LocalFilePrefix) {
		return nil, errors.New("本地文件请以 \"" + prefix + "\" 作为前缀")
	}
	m3u8Url = m3u8Url[7:]
	stat, err := os.Stat(m3u8Url)
	for err != nil || stat.IsDir() {
		// 如果有错误，说明文件不存在，重复判断
		mylog.Info("查找不到本地的 m3u8 文件：" + m3u8Url)
		time.Sleep(time.Second * 3)
		stat, err = os.Stat(m3u8Url)
	}
	// 1 读取数据
	mFile, err := os.Open(m3u8Url)
	if err != nil {
		return nil, errors.Wrapf(err, "打开本地 m3u8 文件出现异常，path: %v", m3u8Url)
	}
	defer mFile.Close()
	ans := []*TsMeta{}
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
		ans = append(ans, &TsMeta{Url: line, Index: len(ans) + 1})
	}
	// 3 删除 m3u8 文件
	err = os.Remove(m3u8Url)
	if err != nil {
		mylog.Warn("删除本地 m3u8 文件失败：" + err.Error())
	}
	return ans, nil
}

// 读取网络 M3U8 文件
// @param m3u8Url url
// @param headers 请求头
// @return ts urls
func readHttpTsUrls(m3u8Url string, headers map[string]string) ([]*TsMeta, error) {
	if !CheckM3U8(m3u8Url, headers) {
		return nil, errors.New("不是规范的 m3u8 地址")
	}

	// 1 找到后缀的位置
	queryPos := strings.Index(m3u8Url, "?")
	if queryPos == -1 {
		queryPos = len(m3u8Url)
	}

	// 2 找到后缀之前的第一个 '/'
	lastSepPos := strings.LastIndex(m3u8Url[:queryPos], "/")
	if lastSepPos == -1 {
		return nil, errors.New("m3u8 地址不规范")
	}
	baseUrl := m3u8Url[:lastSepPos]

	// 3 读取 m3u8 信息
	client := myhttp.TimeoutHttpClient()
	for {
		req, err := http.NewRequest(http.MethodGet, m3u8Url, nil)
		if err != nil {
			util.PrintRetryError("构造请求时发生异常", err, 2)
			continue
		}
		// 添加请求头
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			util.PrintRetryError("发送请求时出现异常", err, 2)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			util.PrintRetryError("发送请求时出现异常", errors.New("错误码："+strconv.Itoa(resp.StatusCode)), 2)
			continue
		}
		defer resp.Body.Close()

		// 逐行扫描 m3u8 文件，将 ts 分片封装成 meta 对象
		scanner := bufio.NewScanner(resp.Body)
		ans := []*TsMeta{}
		xMapUrl := ""
		for scanner.Scan() {
			mt := TsMeta{Index: len(ans) + 1}
			line := scanner.Text()

			// 判断是否是 X-MAP Head 头
			if strings.HasPrefix(line, ExtXMap) {
				if hi, err := ResolveXMap(line); err == nil {
					xMapUrl = baseUrl + "/" + hi.Uri
				}
			}
			mt.HeadUrl = xMapUrl

			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				// 去除注释和空行
				continue
			}

			if !strings.HasPrefix(line, NetworkLinkPrefix) {
				// 补充 baseUrl
				line = baseUrl + "/" + line
			}
			mt.Url = line

			ans = append(ans, &mt)
		}

		return ans, nil
	}
}

// 合并 ts 文件列表
// @param tsDirPath 临时目录
func Merge(tsDirPath string, dmt *meta.Download) error {
	if dmt == nil {
		return errors.New("下载元数据为空")
	}

	dirName := filepath.Base(tsDirPath)
	fileName := dirName[:len(dirName)-len(config.G.Downloader.TsDirSuffix)-1]
	mylog.Infof("准备将 ts 文件合并成 mp4 文件，目标视频：%s", fileName)
	err := transfer.Instance(dmt.OriginUrl).Ts2Mp4(tsDirPath, filepath.Dir(tsDirPath)+"/"+fileName, dmt.LogBar)
	if err != nil {
		return errors.Wrap(err, "合并失败")
	}
	if err = os.RemoveAll(tsDirPath); err != nil {
		mylog.Errorf("临时目录删除失败，目标视频：%s", fileName)
	}
	mylog.Successf("合并完成，目标视频：%s", fileName)
	return nil
}
