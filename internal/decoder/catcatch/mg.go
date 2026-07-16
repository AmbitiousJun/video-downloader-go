package catcatch

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
)

// format2PayloadMg 视频格式转换成分辨率标签
var format2PayloadMg = map[string]string{
	"480P":     "480P",
	"576P":     "576P",
	"720P":     "720P",
	"1080PVIP": "1080P SDR",
}

// MgDecoder 适配芒果 TV 的猫抓解析器, id => cat-catch:mg
// 实现解析器接口
type MgDecoder struct {
	baseDecoder
	// url 记录当前正在解析的 url 地址
	url string
}

// 解析并获取下载链接列表
// 对于 youtube-dl 解析器, 有可能会返回两条链接, 因为部分站点的解析结果是音视频分开的
// 其他解析器通常只返回一条链接
func (mg *MgDecoder) FetchDownloadLinks(url string) ([]string, error) {
	// 解析配置
	mg.url = url
	videoFormat := config.G.Decoder.CatCatch.Sites.Mg.VideoFormat
	if videoFormat == "" {
		return nil, errors.New("未配置要解析的清晰度")
	}
	formatPayload, ok := format2PayloadMg[videoFormat]
	if !ok {
		return nil, fmt.Errorf("错误的清晰度配置: %s", videoFormat)
	}

	// 初始化猫抓解析器
	cc, err := NewCatCather()
	if err != nil {
		return nil, fmt.Errorf("初始化猫抓解析器失败: %v", err)
	}
	defer cc.Close()

	// 读取 Cookie
	mylog.Info("正在读取 Cookie...")
	cookies := mg.ReadCookiesFromConfig(config.G.Decoder.CatCatch.Sites.Mg.CookieJsonPath)

	var nodes []*cdp.Node
	var text string
	var evalRes []string
	err = cc.Run(
		// 跳转到待解析的 url 地址
		chromedp.ActionFunc(func(ctx context.Context) error {
			mylog.Infof("访问待解析地址: %s", url)
			return nil
		}),
		chromedp.Navigate(url),

		// 注入 Cookie
		mg.SetCookiesActionFunc(cookies, url),
		chromedp.Reload(),

		// 通过检查用户头像判断用户是否登录
		chromedp.Nodes(".top-header-v2-actions__avatar__img", &nodes, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()

			tipOnce := sync.OnceFunc(func() {
				mylog.Info("⌛ 等待登录态恢复完成...")
			})

			for {
				select {
				case <-timeoutCtx.Done():
					return errors.New("登录态恢复超时")
				default:
					err := chromedp.Run(ctx, chromedp.Evaluate(`Array.from(document.querySelectorAll('.top-header-v2-actions__avatar__img')).map(e => e.getAttribute('src'))`, &evalRes))
					if err != nil {
						return fmt.Errorf("检验用户登录态失败: %v", err)
					}

					if len(evalRes) <= 0 {
						tipOnce()
						time.Sleep(100 * time.Millisecond)
						continue
					}

					mylog.Successf("成功识别用户登录状态, 头像 url: %s", evalRes[0])
					return nil
				}
			}
		}),

		chromedp.ActionFunc(func(ctx context.Context) error {
			mylog.Infof("⌛ 等待播放开始...")
			return nil
		}),
		chromedp.WaitVisible(".icon .n-pause", chromedp.ByQuery),

		// 注入辅助脚本
		chromedp.ActionFunc(func(ctx context.Context) error {
			mylog.Info("正在注入辅助脚本...")
			return nil
		}),
		mg.AddVideoFormat2ButtonAttribute(&text),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if text != "" {
				return fmt.Errorf("辅助脚本执行异常: %v", text)
			}
			mylog.Success("辅助脚本注入成功")
			return nil
		}),

		// 切换清晰度
		chromedp.Evaluate(fmt.Sprintf(`
			(() => {
				const btn = document.querySelector('[data-value="%s"]');
				if (!btn) {
					return;
				}
				btn.click();
			})()
		`, videoFormat), nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var value string
			var tip string
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*20)
			defer cancel()

			tipOnce := sync.OnceFunc(func() {
				mylog.Infof("⌛ 正在切换至目标清晰度 [%s], 请稍后...", formatPayload)
			})

			for {
				select {
				case <-timeoutCtx.Done():
					return fmt.Errorf("无法将清晰度切换为: %v", formatPayload)
				default:
					err := chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector("._Button_1qs9l_1").innerText`, &value))
					if err != nil {
						return fmt.Errorf("无法获取当前的清晰度: %v", err)
					}

					err = chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector(".bottomLeftTips")?.innerText ?? ""`, &tip))
					if err != nil {
						return fmt.Errorf("检测清晰度切换进度失败: %v", err)
					}
					if !strings.Contains(tip, "已为您切换至") && !strings.Contains(tip, "已切换至") {
						tipOnce()
						time.Sleep(100 * time.Millisecond)
						continue
					}

					if value == formatPayload {
						mylog.Successf("成功切换到清晰度: %s", value)
						return nil
					}

					tipOnce()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("解析异常: %v", err)
	}

	// 注入猫抓脚本, 获取解析结果
	mylog.Info("注入猫抓脚本, 获取资源...")
	results, err := cc.Catch(
		chromedp.WaitVisible(".icon .n-pause", chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("解析不到任何资源")
	}

	// 系统自动检查结果中是否有默认的 m3u8 链接地址, 有则无需用户手动选择
	if dlUrl, ok := mg.ChooseDefaultResult(results); ok {
		return []string{dlUrl}, nil
	}

	// 阻塞系统日志, 调用选择器, 让用户选择要使用抓取到的哪个资源
	dlUrl, err := NewResultSelector(results).Select()
	if err != nil {
		return nil, errors.Wrap(err, "资源选择失败")
	}

	return []string{dlUrl}, nil
}

// ChooseDefaultResult 从猫抓解析结果中自动识别一条可用的 m3u8 地址
func (mg *MgDecoder) ChooseDefaultResult(results []CatCatchResult) (string, bool) {
	for _, res := range results {
		if strings.Contains(res.Url, ".m3u8") {
			return res.Url, true
		}
	}
	return "", false
}

// AddVideoFormat2ButtonAttribute 将视频格式信息添加到对应的按钮的属性中
func (mg *MgDecoder) AddVideoFormat2ButtonAttribute(err *string) chromedp.EvaluateAction {
	return chromedp.Evaluate(`
		(() => {
			// 1 获取清晰度切换按钮列表
			const btns = Array.from(document.querySelectorAll('._item_z42nn_12'));
			if (btns.length <= 0) {
				return "获取不到清晰度切换按钮列表";
			}
			
			// 2 获取清晰度按钮对应的视频格式
			const formats = btns.map(e => e.querySelector('._barName_z42nn_22').innerText)
			if (formats.length != btns.length) {
				return "获取不到清晰度按钮对应的视频格式";
			}
			
			// 3 将视频格式设置到按钮属性上
			for (const i in btns) {
				btns[i].setAttribute("data-value", formats[i]);
			}
			return ""
		})()

	`, err)

}
