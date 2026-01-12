// 封装 youtube-dl 的 format code 选择
// 当用户没有配置 format code 或配置的 code 都解析失败时
// 就让用户自己选择

package ytdl

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog/color"
	"video-downloader-go/internal/util/mystring"

	"github.com/pkg/errors"
)

const (
	ResultPattern     = `\[info\] Available formats for .*` // 用于匹配 youtube-dl 解析结果的正则表达式
	FormatLinePattern = `\b(\d+(\.\d+)?[kKmMgG]i?[bB]?)\b`  // 格式化解析结果的正则表达式，去除大小信息
	StopInput         = "-1"                                // 当接收到用户的指定输入时，停止解析
)

// CodeSelector 封装 youtube-dl 的解析方法
type CodeSelector struct {
	Url string // 要解析的 url 地址

	formatLines []string // 解析结果行，包含 format code 和视频格式相关信息
}

// NewCodeSelector 初始化一个 format code 选择器
func NewCodeSelector(url string) *CodeSelector {
	return &CodeSelector{Url: url}
}

// UserChoice 记录用户选择
type userChoice struct {
	code   string // format code
	format string // 视频格式
}

type userChoiceMap struct {
	sync.Map
}

// host2UserChoice 保存解析 url 和用户选择结果的映射
var host2UserChoice userChoiceMap

// Store 插入一个键值对到 userChoiceMap
func (flm *userChoiceMap) Store(key string, val []*userChoice) {
	flm.Map.Store(key, val)
}

// Load 从 FormatLineMap 中根据键返回值
func (flm *userChoiceMap) Load(key string) []*userChoice {
	val, ok := flm.Map.Load(key)
	if !ok {
		return []*userChoice{}
	}
	return val.([]*userChoice)
}

// RequestCode 用于调用 youtube-dl 请求视频的 format code 列表
// 并接收用户从控制台中输入的 format code
// 最后封装成 YtDlFormatCode 类型的对象返回
func (cs *CodeSelector) RequestCode() (*config.YtDlFormatCode, error) {
	scanner := bufio.NewScanner(os.Stdin)

out:
	for {
		// 执行命令，获取所有可选的 format code
		fmt.Println(color.ToYellow("正在尝试读取 format code..."))
		if err := cs.ExecuteProcess(); err != nil {
			fmt.Println(color.ToRed(fmt.Sprintf("执行命令失败: %v", err)))
			return nil, err
		}

		// 用户选择
		for {
			fmt.Println(color.ToYellow("！！format code 输入规范：[code] 或者 [code1+code2]（不包含[]）"))
			fmt.Println(color.ToYellow(fmt.Sprintf("！！输入自定义的 format code 进行解析，输入空行可重新读取 code，输入 %s 放弃解析", StopInput)))
			fmt.Println(color.ToYellow("请选择要解析的 format code："))

			// 用户配置了记住视频格式 并且成功匹配上之前记录的视频格式
			if code, ok := cs.UseRememberFormat(); ok {
				fmt.Println()
				fmt.Println(color.ToYellow(fmt.Sprintf("通过上一次的选择格式成功匹配到 code: %v, 3 秒后进行解析...", code.Code)))
				time.Sleep(time.Second * 3)
				return code, nil
			}

			if !scanner.Scan() {
				fmt.Println(color.ToYellow("读取输入失败，重新读取 format code..."))
				continue out
			}
			ip := strings.TrimSpace(scanner.Text())
			if ip == "" {
				fmt.Println(color.ToYellow("重新读取 format code..."))
				continue out
			}

			if StopInput == ip {
				return nil, errors.New("用户放弃手动解析")
			}

			// 格式校验
			codes := strings.Split(ip, "+")
			success, fc := false, config.YtDlFormatCode{Code: ip}
			if len(codes) == 1 {
				success, fc.ExpectedLinkNums = true, 1
			}
			if len(codes) == 2 {
				success, fc.ExpectedLinkNums = true, 2
			}

			if !success {
				fmt.Println(color.ToYellow("输入不合法，请重新输入"))
				continue
			}

			// 保存用户选择的格式
			cs.SaveFormat(codes)
			return &fc, nil
		}
	}
}

// UseRememberFormat 根据之前的解析结果以及用户的选择，匹配出当前的 format code
func (cs *CodeSelector) UseRememberFormat() (*config.YtDlFormatCode, bool) {
	// 判断是否开启了 remember format 功能
	rf := config.G.Decoder.YoutubeDL.CustomRememberFormat(cs.Url)
	if rf != config.YoutubeDlRememberFormatActive {
		return nil, false
	}

	// 解析 host
	u, err := url.Parse(cs.Url)
	if err != nil {
		return nil, false
	}

	// 获取解析结果和上次用户选择
	userChoices := host2UserChoice.Load(u.Host)
	if len(userChoices) == 0 || len(cs.formatLines) == 0 {
		return nil, false
	}

	// 构造 format code
	codeBuilder, codeNum := new(strings.Builder), 0

	for _, choice := range userChoices {
		// 根据用户选择的 format 筛选出相同格式的所有解析码
		sameFormatCodes := make([]string, 0)
		for _, formatLine := range cs.formatLines {
			fls := strings.SplitN(formatLine, " ", 2)
			if choice.format == fls[1] {
				sameFormatCodes = append(sameFormatCodes, fls[0])
			}
		}

		// 没有找到相同格式的 code, 说明匹配失败了
		if len(sameFormatCodes) == 0 {
			continue
		}

		// 如果筛选出来的结果中存在用户选择的 code, 直接选择这个 code, 否则选中最后一个 code (通常质量最高)
		finalCode := sameFormatCodes[len(sameFormatCodes)-1]
		for _, code := range sameFormatCodes {
			if code == choice.code {
				finalCode = code
				break
			}
		}

		// 除了第 1 个 code 之外，其他的 code 之前需要拼接 +
		if codeBuilder.Len() > 0 {
			codeBuilder.WriteString("+")
		}
		codeBuilder.WriteString(finalCode)

		// 计数，状态维护
		codeNum++
	}

	// out:
	// 	for _, formatLine := range cs.formatLines {
	// 		fls := strings.SplitN(formatLine, " ", 2)
	// 		for idx, choice := range userChoices {
	// 			// 当前的 choice 已成功匹配
	// 			if vis[idx] {
	// 				continue
	// 			}

	// 			// 格式不匹配并且 code 也不匹配
	// 			if choice.code != fls[0] && choice.format != fls[1] {
	// 				continue
	// 			}

	// 			// 除了第 1 个 code 之外，其他的 code 之前需要拼接 +
	// 			if codeBuilder.Len() > 0 {
	// 				codeBuilder.WriteString("+")
	// 			}
	// 			codeBuilder.WriteString(fls[0])

	// 			// 计数，状态维护
	// 			codeNum++
	// 			vis[idx] = true
	// 			continue out
	// 		}
	// 	}

	// 匹配个数和用户选择的数量不一致
	if codeNum != len(userChoices) {
		return nil, false
	}

	return &config.YtDlFormatCode{Code: codeBuilder.String(), ExpectedLinkNums: codeNum}, true
}

// SaveFormat 接收并保存用户选择的格式
func (cs *CodeSelector) SaveFormat(codes []string) {
	if len(codes) == 0 || len(cs.formatLines) == 0 {
		return
	}

	// 解析出 url 的 host
	u, err := url.Parse(cs.Url)
	if err != nil || u.Host == "" {
		return
	}

	// 存放最终匹配成功的解析结果
	userChoices := []*userChoice{}

	// 每个 code 匹配一个 format line
out:
	for _, formatLine := range cs.formatLines {
		// 将 format line 分割成 code 和 视频格式
		fls := strings.SplitN(formatLine, " ", 2)

		for _, code := range codes {
			if code != fls[0] {
				continue
			}
			userChoices = append(userChoices, &userChoice{code: fls[0], format: fls[1]})
			continue out
		}
	}

	host2UserChoice.Store(u.Host, userChoices)
}

// ExecuteProcess 方法封装命令行程序的执行过程
func (cs *CodeSelector) ExecuteProcess() error {
	// 封装参数
	commands := []string{
		"-F",
		"--no-playlist",
		cs.Url,
	}

	ccf := config.G.Decoder.YoutubeDL.CustomCookiesFrom(cs.Url)
	if ccf != "" && ccf != config.YoutubeDlCookieNone {
		commands = append(commands, "--cookies-from-browser", ccf)
	}

	// 执行命令
	cmd := exec.Command(config.YoutubeDlPath, commands...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行命令失败: %v, cmd: [%s %s]", err, config.YoutubeDlPath, strings.Join(commands, " "))
	}

	cs.PrintFormatCodes(mystring.UTF8(string(output)))
	return nil
}

// PrintFormatCodes 分析 youtube-dl 的解析结果，并输出到控制台中
func (cs *CodeSelector) PrintFormatCodes(raw string) {
	fmt.Println(color.ToYellow("===== 请手动选择 format code"))
	fmt.Println(color.ToYellow("===== 解析地址：" + cs.Url))

	flag := false
	warnPrefix, errPrefix := "WARNING", "ERROR"
	scanner := bufio.NewScanner(strings.NewReader(raw))
	regex := regexp.MustCompile(ResultPattern)
	validFormatLines := []string{}
	for scanner.Scan() {
		res := scanner.Text()
		if strings.HasPrefix(res, warnPrefix) || strings.HasPrefix(res, errPrefix) {
			fmt.Println(color.ToRed(res))
		}
		if !flag && regex.MatchString(res) {
			flag = true
			continue
		}
		if flag {
			fmt.Println(color.ToGreen(res))
			// 把解析结果原始行保存起来
			if tfl, ok := cs.TransferFormatLine(res); ok {
				validFormatLines = append(validFormatLines, tfl)
			}
		}
	}

	cs.formatLines = validFormatLines

	fmt.Println(color.ToYellow("===== 解析地址：" + cs.Url))
}

// TransferFormatLine 负责去掉 format line 中的文件大小信息以及多余的空格
func (cs *CodeSelector) TransferFormatLine(line string) (string, bool) {
	// 忽略起始行
	if strings.HasPrefix(line, "ID") || strings.HasPrefix(line, "-") {
		return "", false
	}

	res := new(strings.Builder)

	// 格式化
	reg := regexp.MustCompile(FormatLinePattern)
	matches := reg.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		line = strings.ReplaceAll(line, match[1], "")
	}

	// 提取 format code
	fields := strings.Fields(line)
	if len(fields) < 1 {
		return "", false
	}
	res.WriteString(fields[0])
	res.WriteString(" ")

	// 拼接后缀
	res.WriteString(strings.Join(fields[1:], ","))
	return res.String(), true
}
