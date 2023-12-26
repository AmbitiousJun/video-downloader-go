// 封装 youtube-dl 的 format code 选择
// 当用户没有配置 format code 或配置的 code 都解析失败时
// 就让用户自己选择

package ytdl

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog"

	"github.com/pkg/errors"
)

const (
	ResultPattern = `\[info\] Available formats for .*` // 用于匹配 youtube-dl 解析结果的正则表达式
	StopInput     = "-1"                                // 当接收到用户的指定输入时，停止解析
)

// CodeSelector 封装 youtube-dl 的解析方法
type CodeSelector struct {
	Url string
}

// NewCodeSelector 初始化一个 format code 选择器
func NewCodeSelector(url string) *CodeSelector {
	return &CodeSelector{Url: url}
}

// RequestCode 用于调用 youtube-dl 请求视频的 format code 列表
// 并接收用户从控制台中输入的 format code
// 最后封装成 YtDlFormatCode 类型的对象返回
func (cs *CodeSelector) RequestCode() (*config.YtDlFormatCode, error) {
	// 在读取的时候，停止日志包输出
	mylog.Block()
	defer mylog.UnBlock()
	scanner := bufio.NewScanner(os.Stdin)

out:
	for {
		// 执行命令，获取所有可选的 format code
		log.Println(mylog.PackMsg("", mylog.ANSIWarning, "正在尝试读取 format code..."))
		if err := cs.ExecuteProcess(); err != nil {
			log.Println(mylog.PackMsg("", mylog.ANSIDanger, fmt.Sprintf("执行命令失败: %v，两秒后重试", err)))
			time.Sleep(time.Second * 2)
			continue
		}

		// 用户选择
		for {
			log.Println(mylog.PackMsg("", mylog.ANSIWarning, "！！format code 输入规范：[code] 或者 [code1+code2]（不包含[]）"))
			log.Println(mylog.PackMsg("", mylog.ANSIWarning, fmt.Sprintf("！！输入自定义的 format code 进行解析，输入空行可重新读取 code，输入 %s 放弃解析", StopInput)))
			log.Println(mylog.PackMsg("", mylog.ANSIWarning, "请选择要解析的 format code："))

			if !scanner.Scan() {
				log.Println(mylog.PackMsg("", mylog.ANSIWarning, "读取输入失败，重新读取 format code..."))
				continue out
			}
			ip := strings.TrimSpace(scanner.Text())
			if ip == "" {
				log.Println(mylog.PackMsg("", mylog.ANSIWarning, "重新读取 format code..."))
				continue out
			}

			if StopInput == ip {
				return nil, errors.New("用户放弃手动解析")
			}

			// 格式校验
			codes := strings.Split(ip, "+")
			if len(codes) == 1 {
				return &config.YtDlFormatCode{Code: ip, ExpectedLinkNums: 1}, nil
			}
			if len(codes) == 2 {
				return &config.YtDlFormatCode{Code: ip, ExpectedLinkNums: 2}, nil
			}

			log.Println(mylog.PackMsg("", mylog.ANSIWarning, "输入不合法，请重新输入"))
		}
	}
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
	if ccf != "" {
		commands = append(commands, "--cookies-from-browser", ccf)
	}

	// 执行命令
	cmd := exec.Command(config.YoutubeDlPath, commands...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "执行命令失败")
	}

	cs.PrintFormatCodes(string(output))
	return nil
}

// PrintFormatCodes 分析 youtube-dl 的解析结果，并输出到控制台中
func (cs *CodeSelector) PrintFormatCodes(raw string) {
	log.Println(mylog.PackMsg("", mylog.ANSIWarning, "===== 请手动选择 format code"))
	log.Println(mylog.PackMsg("", mylog.ANSIWarning, "===== 解析地址："+cs.Url))

	flag := false
	warnPrefix, errPrefix := "WARNING", "ERROR"
	scanner := bufio.NewScanner(strings.NewReader(raw))
	regex := regexp.MustCompile(ResultPattern)
	for scanner.Scan() {
		res := scanner.Text()
		if strings.HasPrefix(res, warnPrefix) || strings.HasPrefix(res, errPrefix) {
			log.Println(mylog.PackMsg("", mylog.ANSIDanger, res))
		}
		if !flag && regex.MatchString(res) {
			flag = true
			continue
		}
		if flag {
			log.Println(mylog.PackMsg("", mylog.ANSISuccess, res))
		}
	}

	log.Println(mylog.PackMsg("", mylog.ANSIWarning, "===== 解析地址："+cs.Url))
}
