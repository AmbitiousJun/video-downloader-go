package decoder

import (
	"fmt"
	"log"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"

	"github.com/pkg/errors"
)

const (
	// 一个 format code 最多可以重试 5 次
	RetryTime = 5
)

// YtDlDecoder 是一个使用 youtube-dl 工具来进行解析的解析器
type YtDlDecoder struct{}

// FetchDownloadLinks 是核心解析方法，实现接口 D
func (ydd *YtDlDecoder) FetchDownloadLinks(url string) ([]string, error) {
	codes := config.G.Decoder.YoutubeDL.FormatCodes

	// 1 尝试配置文件中配置的 format
	if links, err := ydd.tryLinks(url, codes); err == nil {
		return links, nil
	}

	// 2 尝试用户手动输入的 format
	log.Println(mylog.PackMsg("", mylog.ANSIWarning, "预置 code 全部解析失败或没有配置，触发手动选择，url："+url))

	// TODO: 调用 selector 请求 format code
	code := &config.YtDlFormatCode{}

	links, err := ydd.tryLinks(url, []*config.YtDlFormatCode{code})
	if err != nil {
		mylog.Error(fmt.Sprintf("解析失败，地址：%s", url))
		return nil, errors.Wrap(err, "解析失败")
	}
	return links, nil
}

// tryLinks 负责解析下载链接，并允许重试 RetryTime 次
func (ydd *YtDlDecoder) tryLinks(url string, codes []*config.YtDlFormatCode) ([]string, error) {
	for _, code := range codes {
		currentTry := 1
		for currentTry <= RetryTime {
			mylog.Info(fmt.Sprintf("尝试解析地址：%s, format code: %s, 第 %d 次尝试...", url, code.Code, currentTry))
			links, err := NewYtDlHandler(url, code).GetLinks()
			if err == nil {
				return links, nil
			}
			currentTry++
			time.Sleep(time.Second)
		}
	}
	return nil, errors.New("解析失败")
}
