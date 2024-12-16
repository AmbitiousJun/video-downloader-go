package https

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var client *http.Client

// RedirectCodes 有重定向含义的 http 响应码
var RedirectCodes = [4]int{http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect}

func init() {
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// IsSuccessCode 判断 http code 是否为成功状态
func IsSuccessCode(code int) bool {
	codeStr := strconv.Itoa(code)
	return strings.HasPrefix(codeStr, "2")
}

// IsRedirectCode 判断 http code 是否是重定向
//
// 301, 302, 307, 308
func IsRedirectCode(code int) bool {
	for _, valid := range RedirectCodes {
		if code == valid {
			return true
		}
	}
	return false
}

// MapBody 将 map 转换为 ReadCloser 流
func MapBody(body map[string]interface{}) io.ReadCloser {
	if body == nil {
		return nil
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		log.Printf("MapBody 转换失败, body: %v, err : %v", body, err)
		return nil
	}
	return io.NopCloser(bytes.NewBuffer(bodyBytes))
}

// Request 发起 http 请求获取响应
//
// 如果一个请求有多次重定向并且进行了 autoRedirect,
// 则最后一次重定向的 url 会作为第一个参数返回
func Request(method, url string, header http.Header, body io.ReadCloser, autoRedirect bool) (string, *http.Response, error) {
	// 1 转换请求
	var bodyBytes []byte
	if body != nil {
		var err error
		if bodyBytes, err = io.ReadAll(body); err != nil {
			return "", nil, fmt.Errorf("读取请求体失败: %v", err)
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header = header

	// 2 发出请求
	resp, err := client.Do(req)
	if err != nil {
		return url, resp, err
	}

	// 3 对重定向响应的处理
	if autoRedirect && IsRedirectCode(resp.StatusCode) {
		loc := resp.Header.Get("Location")
		if strings.HasPrefix(loc, "/") {
			// 需要拼接上当前请求的前缀后再进行重定向
			loc = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, loc)
		}
		return Request(method, loc, header, body, autoRedirect)
	}
	return url, resp, err
}
