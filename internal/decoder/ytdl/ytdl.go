package ytdl

import (
	"fmt"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mylog/color"

	"github.com/pkg/errors"
)

const (
	// 一个 format code 最多可以重试 3 次
	RetryTime = 3
)

// Decoder 是一个使用 youtube-dl 工具来进行解析的解析器
type Decoder struct{}

// FetchDownloadLinks 是核心解析方法，实现接口 D
func (d *Decoder) FetchDownloadLinks(url string) ([]string, error) {
	codes := config.G.Decoder.YoutubeDL.CustomFormatCodes(url)
	// 1 尝试配置文件中配置的 format
	if links, err := d.tryLinks(url, codes); err == nil {
		return links, nil
	}

	// 2 尝试用户手动输入的 format
	mylog.BlockPanel()
	fmt.Println(color.ToYellow("预置 code 全部解析失败或没有配置，触发手动选择，url：" + url))

	// 3 调用 selector 请求 format code
	code, err := d.tryCode(url)
	mylog.UnBlockPanel()
	if err != nil {
		return nil, err
	}

	// 4 使用获取到的 format code 请求视频下载地址
	links, err := d.tryLinks(url, []*config.YtDlFormatCode{code})
	if err != nil {
		return nil, err
	}
	return links, nil
}

// tryCode 调用 youtube-dl 获取 format code, 并允许重试 RetryTime 次
func (d *Decoder) tryCode(url string) (*config.YtDlFormatCode, error) {
	currentTry := 1
	cs := NewCodeSelector(url)
	for currentTry <= RetryTime {
		code, err := cs.RequestCode()
		if err == nil {
			return code, nil
		}
		currentTry++
		time.Sleep(time.Second)
	}
	return nil, errors.New("获取 format code 失败")
}

// tryLinks 负责解析下载链接，并允许重试 RetryTime 次
func (d *Decoder) tryLinks(url string, codes []*config.YtDlFormatCode) ([]string, error) {
	for _, code := range codes {
		currentTry := 1
		for currentTry <= RetryTime {
			mylog.Infof("尝试解析地址：%s, format code: %s, 第 %d 次尝试...", url, code.Code, currentTry)
			links, err := NewHandler(url, code).GetLinks()
			if err == nil {
				return links, nil
			}
			currentTry++
			time.Sleep(time.Second)
		}
	}
	return nil, errors.New("解析失败")
}
