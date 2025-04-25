package mylog

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"video-downloader-go/internal/util/mylog/color"
	"video-downloader-go/internal/util/mylog/dlbar"
	"video-downloader-go/internal/util/mystring"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// Panel 是全局唯一日志面板
type Panel struct {
	// 日志更新同步
	mu sync.Mutex

	// 日志列表
	dlBars []*dlbar.Bar

	// 存放 barId 和 bar 的映射关系, 方便外部快速取出 bar 对象
	barMap map[string]*dlbar.Bar

	// 当前的下载速率
	dlSpeed func() string

	// 日志列表
	logs []string

	// maxLogNum 面板中最大的日志条数
	maxLogNum int

	// 是否可以注册新的 bar 到当前面板中
	canRegister bool

	// 日志面板的总行数
	totalLine int

	// 是否至少打印过 1 次面板了
	hasPrintAtLeastOnce bool
}

// NewPanel 创建一个日志面板对象
func NewPanel(dlSpeedGetter func() string, maxLogNum int) *Panel {
	p := &Panel{
		dlBars:      []*dlbar.Bar{},
		dlSpeed:     dlSpeedGetter,
		barMap:      make(map[string]*dlbar.Bar),
		logs:        []string{},
		maxLogNum:   maxLogNum,
		canRegister: true,
	}
	for i := 1; i <= p.maxLogNum; i++ {
		p.logs = append(p.logs, "")
	}
	return p
}

// PreventRegister 阻止注册新的 bar 到面板中
func (p *Panel) PreventRegister() {
	p.canRegister = false
	p.CalcTotalLine()
}

// PrintLogPanel 将日志面板输出到控制台上
func (p *Panel) PrintLogPanel(canClear bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	width, _ := p.GetTerminalSize()
	tipsLen := 11
	if width <= tipsLen {
		width = 1
	} else {
		width -= tipsLen
	}

	// 1 将当次面板日志全部收集到数组中
	allLogs := []string{}
	allLogs = append(allLogs, color.ToPurple("Download ↓ "+strings.Repeat("=", width)))
	p.SortBarsByStatus()
	p.CollectBarLogs(func(s string) {
		allLogs = append(allLogs, s)
	})
	allLogs = append(allLogs, color.ToYellow("Speed: "+p.dlSpeed()))
	allLogs = append(allLogs, color.ToPurple("Download ↑ "+strings.Repeat("=", width)))
	allLogs = append(allLogs, color.ToPurple("Log      ↓ "+strings.Repeat("=", width)))
	for _, l := range p.logs {
		allLogs = append(allLogs, fmt.Sprintf("○ %s", l))
	}
	allLogs = append(allLogs, color.ToPurple("Log      ↑ "+strings.Repeat("=", width)))

	// 2 清空旧日志
	if p.hasPrintAtLeastOnce && canClear {
		p.ClearOldPanel()
	}

	// 3 将收集好的日志输出到控制台上
	for _, l := range allLogs {
		fmt.Println(l)
	}
	p.hasPrintAtLeastOnce = true
}

// ClearOldPanel 清空旧日志, 即清掉 TotalLine 行日志
func (p *Panel) ClearOldPanel() {
	for i := 1; i <= p.totalLine; i++ {
		fmt.Print("\033[1A")
		fmt.Print("\033[K")
	}
}

// CalcTotalLine 计算当前输出日志面板需要占用多少行
func (p *Panel) CalcTotalLine() {
	p.totalLine = 5 + len(p.dlBars) + p.maxLogNum
}

// SortBarsByStatus 将 bars 数组根据 status 进行排序
func (p *Panel) SortBarsByStatus() {
	sort.SliceStable(p.dlBars, func(i, j int) bool {
		if p.dlBars[i].Status != p.dlBars[j].Status {
			return p.dlBars[i].Status < p.dlBars[j].Status
		}
		if p.dlBars[i].Status != dlbar.BarStatusExecuting {
			return p.dlBars[i].Name < p.dlBars[j].Name
		}
		if p.dlBars[i].ChildStatus == p.dlBars[j].ChildStatus {
			return p.dlBars[i].Name < p.dlBars[j].Name
		}
		return p.dlBars[i].ChildStatus < p.dlBars[j].ChildStatus
	})
}

// RegisterBar 将一个 bar 对象注册到日志面板中
func (p *Panel) RegisterBar(b *dlbar.Bar) {
	if !p.canRegister || b == nil || b.Id == "" {
		return
	}
	if _, ok := p.barMap[b.Id]; ok {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dlBars = append(p.dlBars, b)
	p.barMap[b.Id] = b
}

// GetBar 获取面板中的 Bar 对象
func (p *Panel) GetBar(barId string) (*dlbar.Bar, error) {
	var res *dlbar.Bar
	var ok bool
	if res, ok = p.barMap[barId]; !ok {
		return nil, errors.New("bar 不存在")
	}
	return res, nil
}

// GetTerminalSize 返回用户运行的终端大小
func (p *Panel) GetTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 160, 90
	}
	return width, height
}

// CollectBarLogs 将多个 bar 批量处理为字符串 log 并以回调形式返回
func (p *Panel) CollectBarLogs(callback func(s string)) {
	allGroups := make([][]string, len(p.dlBars))
	maxLen := make(map[int]int)
	// 统计每一个项目的最长长度
	for i, bar := range p.dlBars {
		g := bar.Group()
		// 在最前面拼接上 bar 序号
		g = append([]string{fmt.Sprintf("%d.", i+1)}, g...)
		allGroups[i] = g
		for j, s := range g {
			sLen := runewidth.StringWidth(s)
			if sLen > maxLen[j] {
				maxLen[j] = sLen
			}
		}
	}
	// 遍历所有分组, 按照最长长度进行格式化返回
	for _, group := range allGroups {
		sb := strings.Builder{}
		for i, s := range group {
			sb.WriteString(mystring.PadRightByRuneWidth(s, maxLen[i], ' '))
			if i < len(group)-1 {
				sb.WriteString(" ")
			}
		}
		callback(sb.String())
	}
}

// AppendLog 追加日志到面板中
//
// 自动截断面板的过时日志
func (p *Panel) AppendLog(l string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logs = append(p.logs, l)
	overflowNum := len(p.logs) - p.maxLogNum
	if overflowNum > 0 {
		p.logs = p.logs[overflowNum:]
	}
}
