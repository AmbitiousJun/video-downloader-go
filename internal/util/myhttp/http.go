package myhttp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"video-downloader-go/internal/util"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mytokenbucket"

	"github.com/pkg/errors"
)

const (
	ConnectTimeout = 60 * time.Second  // 建立连接的超时时间
	ReadTimeout    = 300 * time.Second // 读取数据的超时时间
)

const (
	HttpHeaderRangesPattern     = "bytes=(\\d*)-(\\d*)"  // 用于匹配出 Http 请求头中的 Ranges 的值
	HttpRespHeaderRangesPattern = "bytes (\\d*)-(\\d*)/" // 用于匹配出 Http 响应头中的 Ranges 值
	HttpHeaderRangesKey         = "Range"                // Range 请求头 key
)

const (
	ErrConnectionReset = "connection reset"
)

// 生成一个带有默认 referer 头的 headerMap
// @param baseMap 已存在的 map，提供了这个值的话，就会在这个 map 上加值
// @param url 需要转换的 url
// @return 添加了 referer 头的 headerMap
func GenDefaultHeaderMapByUrl(baseMap map[string]string, url string) map[string]string {
	if baseMap == nil {
		baseMap = make(map[string]string)
	}
	mg := "mgtv.com"
	bili := "bilivideo"
	if strings.Contains(url, mg) {
		baseMap["Referer"] = "https://" + mg
	}
	if strings.Contains(url, bili) {
		baseMap["Referer"] = "https://bilibili.com"
	}
	return baseMap
}

// 获取一个有超时限制的 http 请求客户端（单例模式）
var TimeoutHttpClient = (func() func() *http.Client {
	var client *http.Client = nil
	var once sync.Once
	return func() *http.Client {
		once.Do(func() {
			transport := http.Transport{
				Dial:                  (&net.Dialer{Timeout: ConnectTimeout}).Dial,
				ResponseHeaderTimeout: ReadTimeout,
			}
			client = &http.Client{Transport: &transport}
		})
		return client
	}
})()

// 下载一个网络资源到本地的文件上，并进行网络限速
// @param request 构造好的请求对象
// @param destPath 要下载到本地文件的绝对路径
// @return 下载成功时，返回下载的字节数，下载失败则返回错误
func DownloadWithRateLimitV2(request *http.Request, destPath string) (int64, error) {
	if request == nil {
		return 0, errors.New("request 对象不能为空")
	}

	// 打开目标文件
	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		if util.IsRetryableError(err) {
			util.PrintRetryError("打开文件 ["+destPath+"] 失败", err, 1)
			return DownloadWithRateLimitV2(request, destPath)
		}
		return 0, fmt.Errorf("打开文件 [%s] 失败: %v", destPath, err)
	}
	defer destFile.Close()

	// 发起请求获取响应
	client := TimeoutHttpClient()
	resp, err := client.Do(request)
	if err != nil {
		if util.IsRetryableError(err) {
			util.PrintRetryError("发送请求失败", err, 2)
			return DownloadWithRateLimitV2(request, destPath)
		}
		return 0, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 非成功响应
	if !Is2xxSuccess(resp.StatusCode) {
		if resp.StatusCode == 416 {
			return 0, errors.New("检测到 416 错误码")
		}
		util.PrintRetryError(fmt.Sprintf("错误码：%v", resp.StatusCode), err, 2)
		return DownloadWithRateLimitV2(request, destPath)
	}

	// 从 Content-Range 响应头中读取出起始字节
	var offset int64
	contentRange := resp.Header.Get("Content-Range")
	reg := regexp.MustCompile(HttpRespHeaderRangesPattern)
	if reg.MatchString(contentRange) {
		start := reg.FindStringSubmatch(contentRange)[1]
		startNum, err := strconv.Atoi(start)
		if err != nil {
			mylog.Warnf("非预期的 start 字节, 原始值: %s", contentRange)
		} else {
			offset = int64(startNum)
		}
	}

	// 通过缓冲区分片读取响应
	maxBufSize := 4096
	reader := bufio.NewReader(resp.Body)

	// 不断获取令牌, 将响应的数据写入文件中
	bucket := mytokenbucket.GlobalBucket
	var n int
	var totalBytes int64
	for {
		bufSize := bucket.TryConsume(int64(maxBufSize))
		if bufSize <= 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		buf := make([]byte, bufSize)

		// 不断重试读取响应
		n, err = reader.Read(buf)
		for err != nil && err != io.EOF {
			if err == io.ErrUnexpectedEOF {
				return 0, fmt.Errorf("读取数据异常: %v", err)
			}

			mylog.Warnf("下载异常: %v, 稍后重试...", err)
			time.Sleep(time.Millisecond * 500)
			n, err = reader.Read(buf)
		}
		eof := err == io.EOF

		// 将读取到的字节写入到文件中
		if n > 0 {
			for {
				_, err := destFile.WriteAt(buf[:n], offset+totalBytes)
				if err != nil {
					mylog.Warnf("写入文件异常: %v, 1s 后重试...", err)
					time.Sleep(time.Second)
					continue
				}
				break
			}
			totalBytes += int64(n)
			bucket.CompleteConsume(int64(n))
		}

		// 如果已经读取到文件末尾, 停止读取
		if eof {
			break
		}
	}

	return totalBytes, nil
}

// 下载一个网络资源到本地的文件上，并进行网络限速
// @param request 构造好的请求对象
// @param destPath 要下载到本地文件的绝对路径
// @return 下载成功时，返回下载的字节数，下载失败则返回错误
func DownloadWithRateLimit(request *http.Request, destPath string) (int64, error) {
	url := request.URL.String()
	method := request.Method
	headers := HttpHeader2Map(request.Header.Clone())
	// 1 预请求，获取要下载文件的总大小
	ranges, err := GetRequestRanges(url, method, headers)
	if err != nil {
		if util.IsRetryableError(err) {
			return DownloadWithRateLimit(request, destPath)
		}
		return -1, errors.Wrap(err, "获取文件下载范围失败")
	}
	RemoveRangeHeader(headers)
	start, end := ranges[0], ranges[1]
	// 2 分割文件进行下载，每次下载前先到令牌桶中获取能够下载的字节数
	bucket := mytokenbucket.GlobalBucket
	client := TimeoutHttpClient()
	dest, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return -1, errors.Wrapf(err, "无法打开或创建目标文件：%s", destPath)
	}
	defer dest.Close()
	for start < end {
		consume := bucket.TryConsume(end - start)
		if consume <= 0 {
			// 抢不到令牌，睡眠一小会，防止过渡消耗系统资源
			time.Sleep(time.Millisecond * 100)
			continue
		}
		rangeHeader := fmt.Sprintf("bytes=%v-%v", start, start+consume)
		// 请求部分资源
		subRequest, err := CloneHttpRequest(request)
		if err != nil {
			util.PrintRetryError("克隆请求对象时出现异常", err, 2)
			continue
		}
		subRequest.Header.Set(HttpHeaderRangesKey, rangeHeader)
		resp, err := client.Do(subRequest)
		if err != nil {
			if !strings.Contains(err.Error(), ErrConnectionReset) {
				util.PrintRetryError("发送请求时出现异常", err, 2)
			}
			continue
		}
		code := resp.StatusCode
		if !Is2xxSuccess(code) {
			if code == 416 {
				return -1, errors.New("检测到 416 错误码")
			}
			util.PrintRetryError(fmt.Sprintf("错误码：%v", code), err, 2)
			continue
		}
		if resp.Body == nil {
			util.PrintRetryError("响应体为空", err, 2)
			continue
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			util.PrintRetryError("读取数据异常", err, 2)
			continue
		}
		// 3 将请求下来的文件分片，定位写入到目标文件中
		dest.Seek(start, 0)
		dest.Write(bodyBytes)
		start += consume
		bucket.CompleteConsume(consume)
		resp.Body.Close()
	}
	return ranges[1] - ranges[0], nil
}

// 克隆 http 请求对象
// @param origin 原始对象
// @return 克隆对象
func CloneHttpRequest(original *http.Request) (*http.Request, error) {
	newUrl, err := url.Parse(original.URL.String())
	if err != nil {
		return nil, err
	}
	newHeader := make(http.Header)
	for key, values := range original.Header {
		newHeader[key] = values
	}
	newRequest := &http.Request{
		Method:        original.Method,
		URL:           newUrl,
		Proto:         original.Proto,
		ProtoMajor:    original.ProtoMajor,
		ProtoMinor:    original.ProtoMinor,
		Header:        newHeader,
		Body:          original.Body,
		ContentLength: original.ContentLength,
		Host:          original.Host,
	}
	return newRequest, nil
}

// 将 http 请求头转换为普通的 map
// @param header 请求头
// @return 普通的 map
func HttpHeader2Map(header http.Header) map[string]string {
	res := make(map[string]string)
	if header == nil {
		return res
	}
	for key, values := range header {
		res[key] = values[len(values)-1]
	}
	return res
}

// 移除 Range 请求头
// @param headers 要移除的 map 对象
func RemoveRangeHeader(headers map[string]string) {
	delete(headers, HttpHeaderRangesKey)
	delete(headers, strings.ToLower(HttpHeaderRangesKey))
}

// 判断一个 http 请求的响应吗是否是 2xx 类型的成功码
// @param code 响应码
// @return 是否 2xx
func Is2xxSuccess(code int) bool {
	codeStr := strconv.Itoa(code)
	return strings.HasPrefix(codeStr, "2")
}

// 下载文件时，可以添加 Range 请求头来请求文件的部分字节
// 本函数返回的是要请求的 url 的字节范围
// 发送 http 请求获取 contentLength，返回值是 [from, contentLength]
// @param url 要请求的目的 url
// @param method 请求方法
// @param headers 请求头
// @param from 作为返回值数组中的第一个值
// @return 字节范围
func GetRequestRangesFrom(url, method string, headers map[string]string, from int64) ([]int64, error) {
	if len(url) == 0 || headers == nil {
		return nil, errors.New("url 和 headers 必传")
	}
	RemoveRangeHeader(headers)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "构造请求失败")
	}
	req.Header.Set("Connection", "Close")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := TimeoutHttpClient().Do(req)
	if err != nil {
		return nil, errors.Wrap(err, util.NetworkError.Error())
	}
	defer resp.Body.Close()
	if !Is2xxSuccess(resp.StatusCode) {
		return nil, errors.New(fmt.Sprintf("连接远程地址失败，错误码：%d", resp.StatusCode))
	}
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return nil, errors.New("无法获取资源的 Content-Length 属性")
	}
	return []int64{from, contentLength}, nil
}

// 下载文件时，可以添加 Range 请求头来请求文件的部分字节
// 本方法返回的是要请求的 url 的字节范围
// 如果 headers 中已经存在 Range 头，直接返回
// 否则发送 http 请求获取 contentLength，返回值是 [0, contentLength]
// @param url 要请求的目的 url
// @param method 请求方法
// @param headers 请求头
// @return 字节范围
func GetRequestRanges(url, method string, headers map[string]string) ([]int64, error) {
	if len(url) == 0 || headers == nil {
		return nil, errors.New("url 和 headers 必传")
	}
	if len(method) == 0 {
		method = http.MethodGet
	}
	ranges, ok := headers[HttpHeaderRangesKey]
	if !ok {
		ranges = headers[strings.ToLower(HttpHeaderRangesKey)]
	}
	if len(ranges) == 0 {
		return GetRequestRangesFrom(url, method, headers, 0)
	}
	regex := regexp.MustCompile(HttpHeaderRangesPattern)
	if !regex.MatchString(ranges) {
		// 请求头不合法，忽略
		return GetRequestRangesFrom(url, method, headers, 0)
	}
	m := regex.FindStringSubmatch(ranges)
	from, to := m[1], m[2]
	// from to 全空，是无效的 Range 头，直接去除
	if len(from) == 0 && len(to) == 0 {
		RemoveRangeHeader(headers)
		return GetRequestRanges(url, method, headers)
	}
	// 有 from 没 to，to 直接取 Content-Length
	if len(from) != 0 && len(to) == 0 {
		if fi, err := strconv.ParseInt(from, 10, 64); err != nil {
			return nil, errors.Wrap(err, "错误的 Range 头")
		} else {
			return GetRequestRangesFrom(url, method, headers, fi)
		}
	}
	// 两者都有，直接返回
	if len(from) == 0 {
		from = "0"
	}
	if fi, err := strconv.ParseInt(from, 10, 64); err != nil {
		return nil, errors.Wrap(err, "错误的 Range 头")
	} else if ti, err := strconv.ParseInt(to, 10, 64); err != nil {
		return nil, errors.Wrap(err, "错误的 Range 头")
	} else {
		return []int64{fi, ti}, nil
	}
}
