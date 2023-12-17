package ytdl

import (
	"bufio"
	"os/exec"
	"strings"
	"video-downloader-go/internal/config"

	"github.com/pkg/errors"
)

type Handler struct {
	cmd        *exec.Cmd // 命令行命令
	eptUrlNums int       // 预期将会解析出来的链接个数
}

// NewHandler 用于创建一个 youtube-dl 的处理器
func NewHandler(url string, formatCode *config.YtDlFormatCode) *Handler {
	ydh := &Handler{eptUrlNums: formatCode.ExpectedLinkNums}

	commands := []string{
		"-f", formatCode.Code,
		url,
		"--get-url",
		"--no-playlist",
	}

	if config.G.Decoder.YoutubeDL.CookiesFrom != "" {
		commands = append(commands, "--cookies-from-browser", config.G.Decoder.YoutubeDL.CookiesFrom)
	}

	ydh.cmd = exec.Command(config.YoutubeDlPath, commands...)
	return ydh
}

// GetLinks 用于获取解析结果，并统一返回整个解析过程的错误
func (ydh *Handler) GetLinks() ([]string, error) {
	output, err := ydh.cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "获取命令行输出失败")
	}

	// 逐行读取链接
	links := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "http") {
			links = append(links, l)
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "读取解析地址的时候出现异常")
	}

	if ydh.eptUrlNums != len(links) {
		return nil, errors.New("解析下载地址失败，与预期地址数不一致")
	}

	return links, nil
}
