// 下载日志输出单独处理
package mylog

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/mattn/go-runewidth"
)

const (
	DlLogStart = " 正在下载 ↓ "
	DlLogEnd   = " 正在下载 ↑ "
)

// DownloadLog 记录一个下载任务的日志输出
type DownloadLog struct {
	id   string     // 通过 id 区分下载日志
	logs []*logItem // 日志队列
}

// downloadLogs 存储当前正在活跃的下载任务
//
// key: id => value: DownloadLog
var downloadLogs map[string]*DownloadLog

// logsMutex 用于保证操作 downloadLogs 时协程安全
var logsMutex sync.Mutex

func init() {
	downloadLogs = make(map[string]*DownloadLog)
}

// NewDownloadLog 创建一个新的下载日志，并加入到下载日志输出队列中
func NewDownloadLog() *DownloadLog {
	id := uuid.New().String()
	dl := DownloadLog{id: id, logs: []*logItem{}}
	addDownloadLog(&dl)
	return &dl
}

// Reset 重置下载日志
func (dl *DownloadLog) Reset() {
	dl.logs = make([]*logItem, 0)
}

// Info 添加一条子日志
func (dl *DownloadLog) Info(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("INFO", ANSIInfo, logMsg))
}

// Warn 添加一条子日志
func (dl *DownloadLog) Warn(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("WARN", ANSIWarning, logMsg))
}

// Success 添加一条子日志
func (dl *DownloadLog) Success(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("SUCCESS", ANSISuccess, logMsg))
}

// Error 添加一条子日志
func (dl *DownloadLog) Error(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("ERROR", ANSIDanger, logMsg))
}

// Progress 添加一条子日志
func (dl *DownloadLog) Progress(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("PROGRESS", ANSIProgress, logMsg))
}

// Infof 添加一条格式化日志
func (dl *DownloadLog) Infof(format string, args ...interface{}) {
	dl.Info(fmt.Sprintf(format, args...))
}

// Warnf 添加一条格式化日志
func (dl *DownloadLog) Warnf(format string, args ...interface{}) {
	dl.Warn(fmt.Sprintf(format, args...))
}

// Successf 添加一条格式化日志
func (dl *DownloadLog) Successf(format string, args ...interface{}) {
	dl.Success(fmt.Sprintf(format, args...))
}

// Errorf 添加一条格式化日志
func (dl *DownloadLog) Errorf(format string, args ...interface{}) {
	dl.Error(fmt.Sprintf(format, args...))
}

func (dl *DownloadLog) Progressf(format string, args ...interface{}) {
	dl.Progress(fmt.Sprintf(format, args...))
}

// Trigger 触发日志输出，此方法会让当前所有活跃的下载任务全都重新输出
func (dl *DownloadLog) Trigger() {
	// 1 检查当前活跃的下载任务中是否有当前任务
	if _, ok := downloadLogs[dl.id]; !ok {
		return
	}

	printDownloadLogs()
}

// Invalidate 移除当前下载任务
func (dl *DownloadLog) Invalidate() {
	removeDownloadLog(dl.id)
}

// printDownloadLogs 输出当前活跃的日志队列到控制台中
func printDownloadLogs() {
	logsMutex.Lock()
	defer logsMutex.Unlock()
	logs := make([]*logItem, 0)
	// 总日志行数 = 起始行 + 结尾行 + 若干分隔行 + 下载日志总行数
	totalLogLines := 2 + (len(downloadLogs) - 1)

	// 获取终端中可以显示文本的最大宽度
	width, _ := GetTerminalSize()
	// 去除输出前缀后行内最多还能显示多少个字符
	validWidth := width - ProgressLogPrefixSize

	// 输出起始日志
	phLeftSize := (validWidth - runewidth.StringWidth(DlLogStart)) / 2
	phRightSize := validWidth - phLeftSize - runewidth.StringWidth(DlLogStart)
	logs = append(logs, PackMsg("PROGRESS", ANSIProgress, fmt.Sprintf("%s%s%s", strings.Repeat("·", phLeftSize), DlLogStart, strings.Repeat("·", phRightSize))))

	// 输出下载日志
	var i int
	for _, log := range downloadLogs {
		logs = append(logs, log.logs...)
		totalLogLines += len(log.logs)
		if i != len(downloadLogs)-1 {
			logs = append(logs, PackMsg("PROGRESS", ANSIProgress, strings.Repeat("·", validWidth)))
		}
		i++
	}

	// 输出结束日志
	logs = append(logs, PackMsg("PROGRESS", ANSIProgress, fmt.Sprintf("%s%s%s", strings.Repeat("·", phLeftSize), DlLogEnd, strings.Repeat("·", phRightSize))))

	// 将日志原子性地输入到日志队列中
	// 判断上一条日志是否是下载日志, 是的话, 需要清除控制台
	lastQueueLog := lq.peekLast()
	if (lastQueueLog == nil && downloadLogFlag.Load()) || (lastQueueLog != nil && isDownloadLogEnd(lastQueueLog)) {
		clearLog := PackMsg("PROGRESS", ANSIProgress, fmt.Sprintf("\033[%dF\033[J\r", totalLogLines+1))
		logs = append([]*logItem{clearLog}, logs...)
	} else {
		emptyLog := PackMsg("PROGRESS", ANSIProgress, "")
		logs = append([]*logItem{emptyLog}, logs...)
	}

	lq.offerLastAll(logs...)
}

// addDownloadLog 添加一条下载日志到 downloadLogs 队列中
func addDownloadLog(dl *DownloadLog) {
	if dl == nil {
		return
	}

	logsMutex.Lock()
	defer logsMutex.Unlock()

	downloadLogs[dl.id] = dl
}

// removeDownloadLog 移除一条下载日志
func removeDownloadLog(id string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	delete(downloadLogs, id)
}

// isDownloadLogEnd 返回日志是否是下载任务日志的最后一条日志
func isDownloadLogEnd(li *logItem) bool {
	if li == nil {
		return false
	}
	return strings.Contains(li.string, DlLogEnd)
}
