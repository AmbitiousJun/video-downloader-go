package myhttp

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	ConnectTimeout = 60 * time.Second  // 建立连接的超时时间
	ReadTimeout    = 300 * time.Second // 读取数据的超时时间
)

const (
	HttpHeaderRangesPattern = "bytes=(\\d*)-(\\d*)" // 用于匹配出 Http 请求头中的 Ranges 的值
	HttpHeaderRangesKey     = "Range"               // Range 请求头 key
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
	var mu sync.Mutex
	return func() *http.Client {
		if client == nil {
			mu.Lock()
			defer mu.Unlock()
			if client == nil {
				transport := http.Transport{
					Dial:                  (&net.Dialer{Timeout: ConnectTimeout}).Dial,
					ResponseHeaderTimeout: ReadTimeout,
				}
				client = &http.Client{Transport: &transport}
				return client
			}
		}
		return client
	}
})()

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
		return nil, errors.Wrap(err, "请求失败")
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
		method = "GET"
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
