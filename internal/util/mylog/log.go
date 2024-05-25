package mylog

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/util/mylog/color"
	"video-downloader-go/internal/util/mylog/dlbar"
	"video-downloader-go/internal/util/mytokenbucket"

	"golang.org/x/term"
)

const (
	PanelMaxLogs = 5 // 面板中最大的日志条数
)

// Panel 是全局唯一日志面板
type Panel struct {
	// 日志更新同步
	Mu sync.Mutex

	// 日志列表
	DlBars []*dlbar.Bar

	// 存放 barId 和 bar 的映射关系, 方便外部快速取出 bar 对象
	BarMap map[string]*dlbar.Bar

	// 当前的下载速率
	DlSpeed func() string

	// 日志列表, 最多 5 条
	Logs []string

	// 是否可以注册新的 bar, 一般调用 Start 之后就禁止注册
	CanRegister bool

	// 日志面板的总行数
	TotalLine int

	// 是否至少打印过 1 次面板了
	HasPrintAtLeastOnce bool
}

// GlobalPanel 全局唯一的日志面板对象
var GlobalPanel *Panel

// BlockFlag 为 true 时, 不输出日志面板
var BlockFlag bool = false

// doNotClear 为 true 时, 在打印面板的时候不清空旧日志
var doNotClear bool = true

func init() {
	GlobalPanel = NewPanel(func() string { return mytokenbucket.GlobalBucket.CurrentRateStr })
}

// NewPanel 创建一个日志面板对象
func NewPanel(DlSpeedGetter func() string) *Panel {
	p := &Panel{
		DlBars:      []*dlbar.Bar{},
		DlSpeed:     DlSpeedGetter,
		BarMap:      make(map[string]*dlbar.Bar),
		Logs:        []string{},
		CanRegister: true,
	}
	for i := 1; i <= PanelMaxLogs; i++ {
		p.Logs = append(p.Logs, "")
	}
	return p
}

// BlockPanel 阻塞日志面板打印
func BlockPanel() {
	BlockFlag = true
	doNotClear = true
}

// UnBlockPanel 取消阻塞日志面板打印
func UnBlockPanel() {
	BlockFlag = false
}

// Start 启动一个协程, 持续监听并输出任务日志
func Start() {
	appctx.WaitGroup().Add(1)
	GlobalPanel.CanRegister = false
	GlobalPanel.CalcTotalLine()
	go func() {
		ctx := appctx.Context()
		defer appctx.WaitGroup().Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// 程序终止信号, 最后输出一次日志, 然后退出
				GlobalPanel.PrintLogPanel()
				return
			case <-ticker.C:
				if BlockFlag {
					continue
				}
				GlobalPanel.PrintLogPanel()
			}
		}
	}()
}

// PrintLogPanel 将日志面板输出到控制台上
func (p *Panel) PrintLogPanel() {
	GlobalPanel.Mu.Lock()
	defer GlobalPanel.Mu.Unlock()
	width, _ := GetTerminalSize()
	if width <= 11 {
		width = 1
	} else {
		width -= 11
	}

	// 1 将当次面板日志全部收集到数组中
	allLogs := []string{}
	allLogs = append(allLogs, color.ToPurple("Download ↓ "+strings.Repeat("=", width)))
	GlobalPanel.SortBarsByStatus()
	for i, b := range GlobalPanel.DlBars {
		allLogs = append(allLogs, fmt.Sprintf("%d. %s", i+1, b.String()))
	}
	allLogs = append(allLogs, color.ToYellow("Speed: "+GlobalPanel.DlSpeed()))
	allLogs = append(allLogs, color.ToPurple("Download ↑ "+strings.Repeat("=", width)))
	allLogs = append(allLogs, color.ToPurple("Log      ↓ "+strings.Repeat("=", width)))
	for _, l := range GlobalPanel.Logs {
		allLogs = append(allLogs, fmt.Sprintf("○ %s", l))
	}
	allLogs = append(allLogs, color.ToPurple("Log      ↑ "+strings.Repeat("=", width)))

	// 2 清空旧日志
	if p.HasPrintAtLeastOnce && !doNotClear {
		p.ClearOldPanel()
	}

	// 3 将收集好的日志输出到控制台上
	for _, l := range allLogs {
		fmt.Println(l)
	}
	p.HasPrintAtLeastOnce = true
	doNotClear = false
}

// ClearOldPanel 清空旧日志, 即清掉 TotalLine 行日志
func (p *Panel) ClearOldPanel() {
	for i := 1; i <= p.TotalLine; i++ {
		fmt.Print("\033[1A")
		fmt.Print("\033[K")
	}
}

// CalcTotalLine 计算当前输出日志面板需要多少行
func (p *Panel) CalcTotalLine() {
	p.TotalLine = 5 + len(p.DlBars) + PanelMaxLogs
}

// SortBarsByStatus 将 bars 数组根据 status 进行排序
func (p *Panel) SortBarsByStatus() {
	sort.SliceStable(p.DlBars, func(i, j int) bool {
		if p.DlBars[i].Status != p.DlBars[j].Status {
			return p.DlBars[i].Status < p.DlBars[j].Status
		}
		if p.DlBars[i].Status != dlbar.BarStatusExecuting {
			return p.DlBars[i].Name < p.DlBars[j].Name
		}
		if p.DlBars[i].ChildStatus == p.DlBars[j].ChildStatus {
			return p.DlBars[i].Name < p.DlBars[j].Name
		}
		return p.DlBars[i].ChildStatus < p.DlBars[j].ChildStatus
	})
}

// RegisterBar 将一个 bar 对象注册到日志面板中
func (p *Panel) RegisterBar(b *dlbar.Bar) {
	if !p.CanRegister || b == nil || b.Id == "" {
		return
	}
	if _, ok := p.BarMap[b.Id]; ok {
		return
	}
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.DlBars = append(p.DlBars, b)
	p.BarMap[b.Id] = b
}

// GetBar 获取面板中的 Bar 对象
func (p *Panel) GetBar(barId string) (*dlbar.Bar, error) {
	var res *dlbar.Bar
	var ok bool
	if res, ok = p.BarMap[barId]; !ok {
		return nil, errors.New("bar 不存在")
	}
	return res, nil
}

// GetTerminalSize 返回用户运行的终端大小
func GetTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 160, 90
	}
	return width, height
}

// Infof 格式化插入一条 Info 日志
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Info 插入一条 Info 日志
func Info(l string) {
	GlobalPanel.Mu.Lock()
	defer GlobalPanel.Mu.Unlock()
	GlobalPanel.cutLogArr()
	GlobalPanel.Logs = append(GlobalPanel.Logs, color.ToBlue(cutLog(l)))
}

// Errorf 格式化插入一条 Error 日志
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// Error 插入一条 Error 日志
func Error(l string) {
	GlobalPanel.Mu.Lock()
	defer GlobalPanel.Mu.Unlock()
	GlobalPanel.cutLogArr()
	GlobalPanel.Logs = append(GlobalPanel.Logs, color.ToRed(cutLog(l)))
}

// Warnf 格式化插入一条 Warnf 日志
func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// Warn 插入一条 Warn 日志
func Warn(l string) {
	GlobalPanel.Mu.Lock()
	defer GlobalPanel.Mu.Unlock()
	GlobalPanel.cutLogArr()
	GlobalPanel.Logs = append(GlobalPanel.Logs, color.ToYellow(cutLog(l)))
}

// Successf 格式化插入一条 Successf 日志
func Successf(format string, args ...interface{}) {
	Success(fmt.Sprintf(format, args...))
}

// Success 插入一条 Success 日志
func Success(l string) {
	GlobalPanel.Mu.Lock()
	defer GlobalPanel.Mu.Unlock()
	GlobalPanel.cutLogArr()
	GlobalPanel.Logs = append(GlobalPanel.Logs, color.ToGreen(cutLog(l)))
}

// cutLog 判断一条日志的长度如果超出一行, 就进行截断
func cutLog(l string) string {
	width, _ := GetTerminalSize()
	percent := 0.7

	// 1 l 长度合法, 无需额外的判断
	if float64(len(l)) <= float64(width)*percent {
		return l
	}

	// 2 对字符串进行截断
	lRunes := []rune(l)
	runeWidth := float64(len(lRunes)) / float64(len(l)) * float64(width)
	cutRuneLen := int(runeWidth * percent)
	return string(lRunes[:cutRuneLen]) + "..."
}

// cutLogArrSize 检查日志数组大小是否达到最大值, 是则去掉第一条日志, 也就是最早的日志
func (p *Panel) cutLogArr() {
	if len(p.Logs) < PanelMaxLogs {
		return
	}
	p.Logs = p.Logs[1:]
}
