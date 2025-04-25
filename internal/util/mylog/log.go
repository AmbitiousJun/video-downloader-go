package mylog

import (
	"fmt"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/util/mylog/color"
	"video-downloader-go/internal/util/mytokenbucket"

	"github.com/mattn/go-runewidth"
)

const (
	PanelMaxLogs         = 5           // 面板中最大的日志条数
	PanelRefreshInterval = time.Second // 日志面板刷新间隔
)

var (

	// GlobalPanel 全局唯一的日志面板对象
	GlobalPanel *Panel

	// blockFlag 为 true 时, 不输出日志面板
	blockFlag bool = false

	// doNotClear 为 true 时, 在打印面板的时候不清空旧日志
	doNotClear bool = true
)

func init() {
	GlobalPanel = NewPanel(func() string { return mytokenbucket.GlobalBucket.CurrentRateStr }, PanelMaxLogs)
}

// BlockPanel 阻塞日志面板打印
func BlockPanel() {
	blockFlag = true
	// 线程睡眠一个刷新周期, 确保日志面板不会中途刷新
	time.Sleep(PanelRefreshInterval)
	doNotClear = true
}

// UnBlockPanel 取消阻塞日志面板打印
func UnBlockPanel() {
	blockFlag = false
}

// Start 启动一个协程, 持续监听并输出任务日志
func Start() {
	appctx.WaitGroup().Add(1)
	GlobalPanel.PreventRegister()
	go func() {
		ctx := appctx.Context()
		defer appctx.WaitGroup().Done()
		ticker := time.NewTicker(PanelRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// 程序终止信号, 最后输出一次日志, 然后退出
				GlobalPanel.PrintLogPanel(!doNotClear)
				return
			case <-ticker.C:
				if blockFlag {
					continue
				}
				GlobalPanel.PrintLogPanel(!doNotClear)
				// 每次正常打印完成后, 下一次打印都需要清空旧日志
				doNotClear = false
			}
		}
	}()
}

// Infof 格式化插入一条 Info 日志
func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

// Info 插入一条 Info 日志
func Info(l string) {
	GlobalPanel.AppendLog(color.ToBlue(cutLog(l)))
}

// Errorf 格式化插入一条 Error 日志
func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}

// Error 插入一条 Error 日志
func Error(l string) {
	GlobalPanel.AppendLog(color.ToRed(cutLog(l)))
}

// Warnf 格式化插入一条 Warnf 日志
func Warnf(format string, args ...any) {
	Warn(fmt.Sprintf(format, args...))
}

// Warn 插入一条 Warn 日志
func Warn(l string) {
	GlobalPanel.AppendLog(color.ToYellow(cutLog(l)))
}

// Successf 格式化插入一条 Successf 日志
func Successf(format string, args ...any) {
	Success(fmt.Sprintf(format, args...))
}

// Success 插入一条 Success 日志
func Success(l string) {
	GlobalPanel.AppendLog(color.ToGreen(cutLog(l)))
}

// cutLog 判断一条日志的长度如果超出一行, 就进行截断
func cutLog(l string) string {
	width, _ := GlobalPanel.GetTerminalSize()
	lWidth := runewidth.StringWidth(l)
	percent := 0.5

	// 1 l 长度合法, 无需额外的判断
	if float64(lWidth) <= float64(width)*percent {
		return l
	}

	// 2 对字符串进行截断
	lRunes := []rune(l)
	runeWidth := float64(len(lRunes)) / float64(lWidth) * float64(width)
	cutRuneLen := int(runeWidth * percent)
	return string(lRunes[:cutRuneLen]) + "..."
}
