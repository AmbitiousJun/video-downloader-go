package catcatch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
)

const (
	scriptBasePath = "lib/js"
)

var (
	catchScriptCache string   // 猫抓脚本缓存
	cdpScriptsCache  []string // cdp 脚本缓存, 访问页面时会按顺序执行
)

// UserCookie 记录浏览器的 http cookie 信息
type UserCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

// loadCatCatchScript 加载猫抓脚本, 有缓存时返回缓存
func loadCatCatchScript() (string, error) {
	if catchScriptCache != "" {
		return catchScriptCache, nil
	}

	scriptBytes, err := os.ReadFile(filepath.Join(scriptBasePath, "cat-catch.js"))
	if err != nil {
		return "", errors.Wrap(err, "读取猫抓脚本文件失败")
	}

	catchScriptCache = string(scriptBytes)
	return catchScriptCache, nil
}

// loadCdpScript2Cache 加载单个 cdp 脚本到缓存中
func loadCdpScript2Cache(elms ...string) error {
	elms = append([]string{scriptBasePath}, elms...)
	bytes, err := os.ReadFile(filepath.Join(elms...))
	if err != nil {
		return err
	}
	cdpScriptsCache = append(cdpScriptsCache, string(bytes))
	return nil
}

// loadCdpScripts 加载页面预加载脚本, 有缓存时返回缓存
func loadCdpScripts() ([]string, error) {
	if len(cdpScriptsCache) != 0 {
		return cdpScriptsCache, nil
	}

	// stealth.min.js 隐藏自动化程序特征
	if err := loadCdpScript2Cache("stealth.min.js"); err != nil {
		return nil, err
	}

	// WebGl 指纹模拟插件
	if err := loadCdpScript2Cache("content_script", "page_context", "inject.js"); err != nil {
		return nil, err
	}
	if err := loadCdpScript2Cache("content_script", "inject.js"); err != nil {
		return nil, err
	}
	if err := loadCdpScript2Cache("background", "config.js"); err != nil {
		return nil, err
	}
	if err := loadCdpScript2Cache("background", "chrome.js"); err != nil {
		return nil, err
	}
	if err := loadCdpScript2Cache("background", "runtime.js"); err != nil {
		return nil, err
	}
	if err := loadCdpScript2Cache("background", "common.js"); err != nil {
		return nil, err
	}

	return cdpScriptsCache, nil
}

// CatCatcher 实现了基本的 chromedriver 操作
//
// 可根据不同站点扩展该结构体, 实现解析, 该结构体不实现解析器接口
type CatCatcher struct {
	browserCtx  context.Context      // chromedriver 上下文
	cancelFuncs []context.CancelFunc // 上下文关闭函数, 按顺序关闭
}

// Run 方法是对 chromedp 原生 Run 方法增强, 将防自动化检测脚本提前注入到页面中
func (cc *CatCatcher) Run(actions ...chromedp.Action) error {
	actions = append([]chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			scripts, err := loadCdpScripts()
			if err != nil {
				return err
			}
			for _, script := range scripts {
				_, err = page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
	}, actions...)
	return chromedp.Run(
		cc.browserCtx,
		actions...,
	)
}

// NewCatCather 初始化并运行一个猫抓 chromedriver 实例
func NewCatCather() (*CatCatcher, error) {
	cc := CatCatcher{
		cancelFuncs: []context.CancelFunc{},
	}

	// chromedriver 启动参数
	opts := append(
		chromedp.DefaultExecAllocatorOptions[3:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 S```afari/537.36"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("use-gl", "angle"),
	)

	if config.G.Decoder.CatCatch.Headless == config.CatCatchHeadlessActive {
		opts = append(opts, chromedp.Headless)
	}

	// 生成 context
	allocator, c := chromedp.NewExecAllocator(context.Background(), opts...)
	defer func() { cc.cancelFuncs = append(cc.cancelFuncs, c) }()
	browserCtx, cancel := chromedp.NewContext(allocator)
	defer func() { cc.cancelFuncs = append(cc.cancelFuncs, cancel) }()
	timeoutCtx, tc := context.WithTimeout(browserCtx, time.Minute*2)
	defer func() { cc.cancelFuncs = append(cc.cancelFuncs, tc) }()

	// 运行一遍, 没报错才能进行后续解析任务
	if err := chromedp.Run(browserCtx); err != nil {
		return nil, err
	}

	cc.browserCtx = timeoutCtx
	return &cc, nil
}

// Close 解析完成后, 需要调用 Close 关闭服务
func (cc *CatCatcher) Close() {
	for _, cancel := range cc.cancelFuncs {
		cancel()
	}
}

// 对猫抓结果的每个字段打印最大长度进行限制, 便于格式化输出
const (
	CCR_OmitLen      = 3 // 省略符长度
	CCR_ActionLen    = 15 + CCR_OmitLen
	CCR_UrlLen       = 30 + CCR_OmitLen
	CCR_HrefLen      = 15 + CCR_OmitLen
	CCR_ExtLen       = 2 + CCR_OmitLen
	CCR_RequestIdLen = 15 + CCR_OmitLen
)

// CatCatchResult 是猫抓脚本抓取资源的封装
type CatCatchResult struct {
	Action    string `json:"action"`
	Url       string `json:"url"`
	Href      string `json:"href"`
	Ext       string `json:"ext"`
	RequestId string `json:"requestId"`
}

// Catch 注入猫抓脚本后将抓取到的资源收集为 JSON 格式并返回
//
// 可传入多个用于判断资源收集结束的 action, 没有传递则默认阻塞 15 秒后返回现有结果
func (cc *CatCatcher) Catch(waitActions ...chromedp.Action) ([]CatCatchResult, error) {
	// 加载猫抓脚本
	catchScript, err := loadCatCatchScript()
	if err != nil {
		return nil, errors.Wrap(err, "加载猫抓脚本失败")
	}

	actions := []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(catchScript).Do(ctx)
			if err != nil {
				mylog.Errorf("猫抓脚本注入失败, err: %v", err)
				return err
			}
			mylog.Success("猫抓脚本注入成功")
			return nil
		}),
		// 重新加载当前页面, 使得猫抓脚本生效
		chromedp.Reload(),
	}

	// 没有传参，使用默认的阻塞 action
	if len(waitActions) == 0 {
		waitActions = append(waitActions, chromedp.Sleep(time.Second*15))
	}
	actions = append(actions, waitActions...)

	// 获取抓取结果 (JSON)
	var catchResult string
	actions = append(actions, chromedp.Text("#cat-catch-result", &catchResult, chromedp.ByQuery))

	err = cc.Run(actions...)
	if err != nil {
		return nil, errors.Wrap(err, "执行自动化脚本异常")
	}

	// 封装数据
	datas := []CatCatchResult{}
	if err = json.Unmarshal([]byte(catchResult), &datas); err != nil {
		return nil, errors.Wrap(err, "无法正常抓取数据: JSON 转换异常")
	}

	return datas, nil
}

// PrintResult 将猫抓结果打印到控制台上
//
// 这个函数只负责格式化, 调用方必须要提供一个处理函数, 来对每一行数据进行操作
func PrintResult(results []CatCatchResult, lineHandler func(line string)) {
	// 第一行属性名
	lineHandler("         RequestId|  Ext|                              Url|            Action|              Href")
	// 第二行分隔符
	lineHandler("------------------|-----|---------------------------------|------------------|------------------")

	// 逐行输出猫抓结果
	if len(results) == 0 {
		return
	}
	for i := 0; i < len(results); i++ {
		line, result := "", results[i]

		requestId := result.RequestId
		if len(requestId) > CCR_RequestIdLen {
			requestId = requestId[:CCR_RequestIdLen-CCR_OmitLen] + "..."
		}
		line += fmt.Sprintf("%"+strconv.Itoa(CCR_RequestIdLen)+"s|", requestId)

		ext := result.Ext
		if len(ext) > CCR_ExtLen {
			ext = ext[:CCR_ExtLen-CCR_OmitLen] + "..."
		}
		line += fmt.Sprintf("%"+strconv.Itoa(CCR_ExtLen)+"s|", ext)

		url := result.Url
		if len(url) > CCR_UrlLen {
			url = url[:CCR_UrlLen-CCR_OmitLen] + "..."
		}
		line += fmt.Sprintf("%"+strconv.Itoa(CCR_UrlLen)+"s|", url)

		action := result.Action
		if len(action) > CCR_ActionLen {
			action = action[:CCR_ActionLen-CCR_OmitLen] + "..."
		}
		line += fmt.Sprintf("%"+strconv.Itoa(CCR_ActionLen)+"s|", action)

		href := result.Href
		if len(href) > CCR_HrefLen {
			href = href[:CCR_HrefLen-CCR_OmitLen] + "..."
		}
		line += fmt.Sprintf("%"+strconv.Itoa(CCR_HrefLen)+"s", href)

		lineHandler(line)
	}
}
