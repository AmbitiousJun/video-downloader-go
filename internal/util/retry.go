// 提供一些重试工具函数
package util

import (
	"strings"
	"time"
	"video-downloader-go/internal/util/mylog"
)

var retryableErrors = []error{
	NetworkError,
	ConnectionResetError,
}

// 判断是否是可重试异常
func IsRetryableError(err error) bool {
	for _, e := range retryableErrors {
		if strings.Contains(err.Error(), e.Error()) {
			return true
		}
	}
	return false
}

// 输出重试错误
func PrintRetryError(prefix string, err error, seconds int64) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	mylog.Warnf("%v：%v，%d 秒后重试", prefix, errMsg, seconds)
	time.Sleep(time.Second * time.Duration(seconds))
}
