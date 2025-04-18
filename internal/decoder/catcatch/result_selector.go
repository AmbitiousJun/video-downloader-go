// 猫抓解析结果选择器
package catcatch

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mylog/color"

	"github.com/pkg/errors"
)

const (
	StopInput = "-1" // 用户取消选择
)

// ResultSelector 是一个负责与用户在控制台进行交互的解析结果选择器
type ResultSelector struct {
	results []CatCatchResult
}

// NewResultSelector 初始化一个猫抓解析结果选择器
func NewResultSelector(results []CatCatchResult) *ResultSelector {
	return &ResultSelector{
		results: results,
	}
}

// Select 阻塞系统日志, 并让用户在控制台中输入要下载哪个资源
func (rs *ResultSelector) Select() (string, error) {
	mylog.BlockPanel()
	defer mylog.UnBlockPanel()
	scanner := bufio.NewScanner(os.Stdin)

	// 输出解析到的资源
	PrintResult(rs.results, func(line string) {
		fmt.Println(color.ToGreen(line))
	})

	// 用户输入
	for {
		fmt.Println(color.ToYellow(fmt.Sprintf("！！输入要下载资源的 RequestId 执行下载, 输入 %s 放弃解析", StopInput)))
		fmt.Println(color.ToYellow("请输入: "))

		if !scanner.Scan() {
			fmt.Println(color.ToRed("读取输入失败, 请重新输入"))
			continue
		}
		ip := strings.TrimSpace(scanner.Text())
		if ip == "" {
			fmt.Println(color.ToRed("读取输入失败, 请重新输入"))
			continue
		}

		if StopInput == ip {
			return "", errors.New("用户放弃解析")
		}

		// 根据 RequestId 获取资源
		url, err := rs.GetUrlByRequestId(ip)
		if err != nil {
			fmt.Println(color.ToRed(fmt.Sprintf("%v, 请重新输入", err)))
			continue
		}

		return url, nil
	}
}

// GetUrlByRequestId 从 results 中匹配 url
//
// 匹配失败时, 会返回 error
func (rs *ResultSelector) GetUrlByRequestId(requestId string) (string, error) {
	for _, result := range rs.results {
		if result.RequestId == requestId {
			return result.Url, nil
		}
	}
	return "", errors.New("RequestId 匹配失败")
}
