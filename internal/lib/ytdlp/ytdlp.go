package ytdlp

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Init 空函数, 目的是触发包内的初始化
func Init() {}

// Extract 使用指定的 formatCode 解析 url
func Extract(url, formatCode string) (string, error) {
	if !execOk {
		return "", errors.New("yt-dlp 环境未初始化")
	}

	// 构造命令
	cmd := exec.Command(
		execPath,
		"-f", formatCode,
		url,
		"--get-url",
	)

	// 执行获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	// 校验结果
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if !scanner.Scan() {
		return "", fmt.Errorf("解析出非预期结果, 原始输出: %s", string(output))
	}

	line := scanner.Text()
	if !strings.HasPrefix(line, "http") {
		return "", fmt.Errorf("解析出非预期结果, 原始输出: %s", string(output))
	}

	return line, nil
}
