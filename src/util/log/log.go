// 日志输出
package log

import (
	"context"
	sysLog "log"
	"math"
	"sync"
	"time"
	"video-downloader-go/src/appctx"
)

// 日志颜色输出常量
const (
	ANSIInfo    = "\x1b[38;2;90;156;248m"
	ANSISuccess = "\x1b[38;2;126;192;80m"
	ANSIWarning = "\x1b[38;2;220;165;80m"
	ANSIDanger  = "\x1b[38;2;228;116;112m"
	ANSIReset   = "\x1b[0m"
)

// 日志队列的最大长度
const logQueueMaxLength int = math.MaxInt32 / 2

// 存放日志的队列
var lq = &logQueue{}

// 日志阻塞标志
var blockFlag = false

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
	lq.offerLast(packMsg("INFO", ANSIInfo, logMsg))
}

// 添加一条 success 日志到队列中
func Success(logMsg string) {
	lq.offerLast(packMsg("SUCCESS", ANSISuccess, logMsg))
}

// 添加一条 error 日志到队列中
func Error(logMsg string) {
	lq.offerLast(packMsg("ERROR", ANSIDanger, logMsg))
}

// 添加一条 warn 日志到队列中
func Warn(logMsg string) {
	lq.offerLast(packMsg("WARN", ANSIWarning, logMsg))
}

// 给日志封装上颜色输出标志
// prefix 日志前缀
// logType 颜色输出标志
// logMsg 要封装的消息
func packMsg(prefix, logType, logMsg string) *logItem {
	return &logItem{logType + prefix + " " + logMsg + ANSIReset}
}

// 输出日志
func printLog(li *logItem) {
	if li == nil {
		return
	}
	sysLog.Println(li.string)
}

// 输出当前队列中的所有日志
func printAllLogs() {
	for HasLog() {
		item := lq.pollFirst()
		printLog(item)
	}
}

// 初始化日治包
func init() {
	logInit := make(chan struct{})
	appctx.WaitGroup().Add(1)
	go listenAndPrintLogs(logInit)
	<-logInit
}

// 初始化日志包
func InitLog(ctx context.Context, wg *sync.WaitGroup) {
	logInit := make(chan struct{})
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
			printAllLogs()
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
