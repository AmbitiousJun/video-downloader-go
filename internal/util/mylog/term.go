// 终端工具函数
package mylog

import (
	"os"
	"sync"

	"golang.org/x/term"
)

const (
	// 日志输出前缀长度：'2023/12/17 01:15:15 SUCCESS '
	SuccessLogPrefixSize = 28
)

// getTerminalSizeErrorOnce 用于控制只输出一次错误信息
var getTerminalSizeErrorOnce sync.Once

// GetTerminalSize 返回用户运行的终端大小
func GetTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		getTerminalSizeErrorOnce.Do(func() {
			Errorf("获取终端大小失败: %v, 使用默认值代替", err)
		})
		width, height = 160, 90
	}
	return width, height
}
