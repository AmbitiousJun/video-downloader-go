// 下载日志输出单独处理
package mylog

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/mattn/go-runewidth"
)

const (
	DlLogStart = " DOWNLOADING ↓ "
	DlLogEnd   = " DOWNLOADING ↑ "
)

// lastPrintLines 记录最后一次输出下载日志时的日志行数
// -1 表示还未输出过下载日志
var lastPrintLines atomic.Int64

// DownloadLog 记录一个下载任务的日志输出
type DownloadLog struct {
	id   string     // 通过 id 区分下载日志
	logs []*logItem // 日志队列

	// 标记当前日志是否准备好输出
	// 队列中只要有一个下载任务未就绪, 整个下载日志就都不会输出
	ready bool
}

// downloadLogs 存储当前正在活跃的下载任务
//
// key: id => value: DownloadLog
var downloadLogs map[string]*DownloadLog

// logsMutex 用于保证操作 downloadLogs 时协程安全
var logsMutex sync.RWMutex

func init() {
	downloadLogs = make(map[string]*DownloadLog)
	lastPrintLines.Store(-1)
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
	dl.ready = false
}

// Info 添加一条子日志
func (dl *DownloadLog) Info(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("INFO", ANSIInfo, logMsg))
	dl.ready = false
}

// Warn 添加一条子日志
func (dl *DownloadLog) Warn(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("WARN", ANSIWarning, logMsg))
	dl.ready = false
}

// Success 添加一条子日志
func (dl *DownloadLog) Success(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("SUCCESS", ANSISuccess, logMsg))
	dl.ready = false
}

// Error 添加一条子日志
func (dl *DownloadLog) Error(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("ERROR", ANSIDanger, logMsg))
	dl.ready = false
}

// Progress 添加一条子日志
func (dl *DownloadLog) Progress(logMsg string) {
	dl.logs = append(dl.logs, PackMsg("PROGRESS", ANSIProgress, logMsg))
	dl.ready = false
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
	// 检查当前活跃的下载任务中是否有当前任务
	if _, ok := downloadLogs[dl.id]; !ok {
		return
	}
	dl.ready = true
	printDownloadLogs()
}

// Invalidate 移除当前下载任务
func (dl *DownloadLog) Invalidate() {
	removeDownloadLog(dl.id)
}

// printDownloadLogs 输出当前活跃的日志队列到控制台中
func printDownloadLogs() {
	logsMutex.RLock()
	defer logsMutex.RUnlock()
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
		// 有日志正在更新, 不进行输出
		if !log.ready {
			return
		}

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
	if needClearConsole(totalLogLines) {
		clearLog := PackMsg("PROGRESS", ANSIProgress, fmt.Sprintf("\033[%dF\033[J\r", totalLogLines+1))
		logs = append([]*logItem{clearLog}, logs...)
	} else {
		emptyLog := PackMsg("PROGRESS", ANSIProgress, "")
		logs = append([]*logItem{emptyLog}, logs...)
	}

	// 更新下载日志行数
	if lastPrintLines.Load() != int64(totalLogLines) {
		lastPrintLines.Store(int64(totalLogLines))
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

// needClearConsole 返回当前是否需要清空控制台的旧下载日志
// 参数 lines 表示当前要输出的下载日志总行数, 如果和上一次输出的下载日志总行数不一致, 则不清空控制台
func needClearConsole(lines int) bool {
	// 当前行数和上一次输出的行数不一致
	lastLines := lastPrintLines.Load()
	if lastLines != -1 && lastLines != int64(lines) {
		return false
	}

	lastQueueLog := lq.peekLast()
	// 当前日志队列为空, 但控制台的最后一条已输出的日志是下载日志
	if lastQueueLog == nil && downloadLogFlag.Load() {
		return true
	}

	// 当前日志队列不为空, 判断最后一条日志是否是下载日志结束
	if lastQueueLog != nil && isDownloadLogEnd(lastQueueLog) {
		return true
	}

	return false
}

// isDownloadLogEnd 返回日志是否是下载任务日志的最后一条日志
func isDownloadLogEnd(li *logItem) bool {
	if li == nil {
		return false
	}
	return strings.Contains(li.string, DlLogEnd)
}
