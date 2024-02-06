// 日志输出
package mylog

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync/atomic"
	"time"
	"video-downloader-go/internal/appctx"
)

// 日志颜色输出常量
const (
	ANSIInfo     = "\x1b[38;2;90;156;248m"
	ANSISuccess  = "\x1b[38;2;126;192;80m"
	ANSIWarning  = "\x1b[38;2;220;165;80m"
	ANSIDanger   = "\x1b[38;2;228;116;112m"
	ANSIProgress = "\x1b[38;2;160;186;250m"
	ANSIReset    = "\x1b[0m"
)

// 日志队列的最大长度
const logQueueMaxLength int = math.MaxInt32 / 2

// 存放日志的队列
var lq = new(logQueue)

// 日志阻塞标志
var blockFlag = false

// 标记最后一次输出日志时是否是下载日志
var downloadLogFlag atomic.Bool

// 判断当前队列中是否还有日志
func HasLog() bool {
	return lq.Len() > 0
}

// 解除日志输出阻塞
func UnBlock() {
	blockFlag = false
}

// 阻塞日志输出
func Block() {
	blockFlag = true
}

// 添加一条 info 日志到队列中
func Info(logMsg string) {
	lq.offerLast(PackMsg("INFO", ANSIInfo, logMsg))
}

// 格式化输出 info 日志
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// 添加一条 success 日志到队列中
func Success(logMsg string) {
	lq.offerLast(PackMsg("SUCCESS", ANSISuccess, logMsg))
}

// 格式化输出 success 日志
func Successf(format string, args ...interface{}) {
	Success(fmt.Sprintf(format, args...))
}

// 添加一条 error 日志到队列中
func Error(logMsg string) {
	lq.offerLast(PackMsg("ERROR", ANSIDanger, logMsg))
}

// 格式化输出一条 error 日志
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// 添加一条 warn 日志到队列中
func Warn(logMsg string) {
	lq.offerLast(PackMsg("WARN", ANSIWarning, logMsg))
}

// 格式化输出一条 warn 日志
func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// 添加一条 progress 日志到队列中
func Progress(logMsg string) {
	lq.offerLast(PackMsg("PROGRESS", ANSIProgress, logMsg))
}

// 格式化输出一条 progress 日志
func Progressf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// 给日志封装上颜色输出标志
// prefix 日志前缀
// logType 颜色输出标志
// logMsg 要封装的消息
func PackMsg(prefix, logType, logMsg string) *logItem {
	if strings.TrimSpace(prefix) == "" {
		return &logItem{string: logType + logMsg + ANSIReset}
	}
	return &logItem{string: logType + prefix + " " + logMsg + ANSIReset}
}

// 输出日志
func printLog(li *logItem) {
	if li == nil {
		return
	}

	// 如果当前日志是下载日志的结尾, 就进行标记
	// 这里无需使用锁同步, 因为系统所有日志都是通过同一个协程逐条输出的, 不存在协程安全问题
	downloadLogFlag.Store(isDownloadLogEnd(li))

	log.Println(li)
}

// 输出当前队列中的所有日志
func PrintAllLogs() {
	for HasLog() {
		item := lq.pollFirst()
		printLog(item)
	}
}

// 初始化日志包
func init() {
	logInit := make(chan struct{})
	appctx.WaitGroup().Add(1)
	go listenAndPrintLogs(logInit)
	<-logInit
}

// 监听日志队列并统一输出日志
func listenAndPrintLogs(logInit chan struct{}) {
	ctx := appctx.Context()
	defer appctx.WaitGroup().Done()
	Info("日志输出线程启动成功，开始监听并输出日志...")
	logInit <- struct{}{}
	for {
		select {
		case <-ctx.Done():
			// 程序终止信号，输出所有的日志
			PrintAllLogs()
			return
		default:
			if !HasLog() || blockFlag {
				// 如果日志阻塞队列长度超出预期的最大值，就抛弃掉旧的日志
				for lq.Len() > logQueueMaxLength {
					lq.pollFirst()
				}
				time.Sleep(time.Second)
				continue
			}
			item := lq.pollFirst()
			printLog(item)
		}
	}
}
