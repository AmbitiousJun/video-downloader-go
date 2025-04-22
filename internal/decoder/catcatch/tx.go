package catcatch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
)

// format2Payload 视频格式转换成分辨率标签
var format2Payload = map[string]string{
	"hd":    "480P",
	"shd":   "720p",
	"fhd":   "1080P",
	"uhd":   "4K",
	"hdr10": "臻彩1080P",
}

// 适配腾讯视频的猫抓解析器, id => cat-catch:tx
// 实现解析器接口
type TxDecoder struct {
	url string // 记录当前正在解析的 url 地址
}

// 解析资源
func (td *TxDecoder) FetchDownloadLinks(url string) ([]string, error) {
	td.url = url
	videoFormat := config.G.Decoder.CatCatch.Sites.Tx.VideoFormat
	if videoFormat == "" {
		return nil, errors.New("未配置要解析的清晰度")
	}
	formatPayload, ok := format2Payload[videoFormat]
	if !ok {
		return nil, fmt.Errorf("错误的清晰度配置: %s", videoFormat)
	}

	cc, err := NewCatCather()
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	// 读取 Cookie
	mylog.Info("正在读取 Cookie ...")
	cookies := td.ReadCookiesFromConfig()

	var nodes []*cdp.Node
	var text string
	err = cc.Run(
		// 跳转到待解析的 url 地址
		chromedp.ActionFunc(func(ctx context.Context) error {
			mylog.Infof("访问待解析地址: %s", url)
			return nil
		}),
		chromedp.Navigate(url),

		// 注入 Cookie
		chromedp.ActionFunc(func(ctx context.Context) error {
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
		}),
		chromedp.Reload(),
		chromedp.WaitVisible(".quick_user_avatar", chromedp.ByQuery),
		chromedp.Sleep(time.Millisecond*100),

		// 通过检查用户头像判断用户是否登录
		chromedp.Nodes(".quick_user_avatar", &nodes, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if len(nodes) == 0 {
				return errors.New("识别不到用户头像")
			}
			src := nodes[0].AttributeValue("src")
			if strings.Contains(src, "common_avatar") {
				return errors.New("恢复登录态失败")
			}
			mylog.Successf("成功识别用户登录状态, 头像 url: %s", src)
			return nil
		}),

		// 注入 JS 脚本, 弹出清晰度选择框
		td.ShowPlayerCover(),
		chromedp.Sleep(time.Second*2),

		// 点击用户指定的清晰度按钮
		chromedp.Click(fmt.Sprintf("[data-value=%s]", videoFormat), chromedp.ByQuery),

		// 等待清晰度切换
		chromedp.Reload(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			mylog.Info("等待页面加载完成...")
			return nil
		}),
		chromedp.WaitVisible("[data-status=pause]", chromedp.ByQuery),

		// 获取当前的清晰度
		td.ShowPlayerCover(),
		chromedp.Text(".txp_btn_definition .txp_label", &text, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if text == "" {
				return errors.New("获取不到视频清晰度")
			}
			if text != formatPayload {
				return fmt.Errorf("获取到的清晰度 [%s] 与预期 [%s] 不一致", text, formatPayload)
			}
			mylog.Successf("成功获取到清晰度: %s", text)
			return nil
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "解析页面时发生错误")
	}

	// 注入猫抓脚本, 获取解析结果
	mylog.Info("注入猫抓脚本, 获取资源...")
	results, err := cc.Catch()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("解析不到任何资源")
	}

	// 系统自动检查结果中是否有默认的 m3u8 链接地址, 有则无需用户手动选择
	if dlUrl, ok := td.ChooseDefaultResult(results); ok {
		return []string{dlUrl}, nil
	}

	// 阻塞系统日志, 调用选择器, 让用户选择要使用抓取到的哪个资源
	dlUrl, err := NewResultSelector(results).Select()
	if err != nil {
		return nil, errors.Wrap(err, "资源选择失败")
	}

	return []string{dlUrl}, nil
}

// ShowPlayerCover 往页面中注入辅助脚本, 使得原本被隐藏的播放器信息能够显示
func (td *TxDecoder) ShowPlayerCover() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		mylog.Info("正在注入辅助 JS 脚本...")
		_, errDts, err := runtime.Evaluate(`
			const hides = document.querySelectorAll('.txp_popup.txp_popup_definition.txp_none, .plugin_ctrl_txp_bottom.txp_none');
			for (let i = 0; i < hides.length; i++) {
				hides[i].classList.remove('txp_none');
			}
		`).Do(ctx)
		if errDts != nil || err != nil {
			mylog.Warnf("注入 JS 脚本失败, 可能导致无法解析到指定清晰度资源, url: %s, errDts: %v, err: %v", td.url, errDts, err)
		}
		return nil
	})
}

// ReadCookiesFromConfig 根据全局配置中的 JSON Cookie 文件路径加载 Cookie 数据
//
// 方法不返回异常, 如果解析出错, 打印日志并直接返回一个空切片
func (td *TxDecoder) ReadCookiesFromConfig() []*UserCookie {
	res := make([]*UserCookie, 0)

	jsonPath := config.G.Decoder.CatCatch.Sites.Tx.CookieJsonPath
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

// ChooseDefaultResult 选中默认的 m3u8 地址
func (td *TxDecoder) ChooseDefaultResult(results []CatCatchResult) (string, bool) {
	dftHost := "apd-vlive.apdcdn.tc.qq.com"
	for _, res := range results {
		if strings.Contains(res.Url, dftHost) {
			return res.Url, true
		}
	}
	return "", false
}
