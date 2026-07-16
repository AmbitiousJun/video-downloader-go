package catcatch

import (
	"context"
	"encoding/json"
	"os"
	"time"
	"video-downloader-go/internal/util/mylog"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// baseDecoder 基础的解析器类型, 提供通用的处理方法
type baseDecoder struct{}

// ReadCookiesFromConfig 根据全局配置中的 JSON Cookie 文件路径加载 Cookie 数据
//
// 方法不返回异常, 如果解析出错, 打印日志并直接返回一个空切片
func (d *baseDecoder) ReadCookiesFromConfig(jsonPath string) []*UserCookie {
	res := make([]*UserCookie, 0)

	if jsonPath == "" {
		return res
	}

	// 读取文件
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		mylog.Warnf("加载 Cookie 异常, 读取文件失败, path: %s, error: %v", jsonPath, err)
		return res
	}

	// 反序列化
	if err = json.Unmarshal(jsonBytes, &res); err != nil {
		mylog.Warnf("加载 Cookie 异常, 反序列化 JSON 失败, path: %s, error: %v", jsonPath, err)
	}

	return res
}

// SetCookiesActionFunc 返回一个设置 Cookie 的 Action
func (d *baseDecoder) SetCookiesActionFunc(cookies []*UserCookie, url string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		mylog.Info("正在注入 Cookie ...")
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
		for _, cookie := range cookies {
			err := network.SetCookie(cookie.Name, cookie.Value).
				WithExpires(&expr).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithSecure(false).
				Do(ctx)
			if err != nil {
				mylog.Warnf("注入 Cookie 失败, url: %s, error: %v", url, err)
				return err
			}
		}
		return nil
	})
}
