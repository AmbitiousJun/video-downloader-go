// 提供一些重试工具函数
package util

import (
	"fmt"
	"strings"
	"time"
	"video-downloader-go/internal/util/mylog"
)

var retryableErrors = map[string]struct{}{
	NetworkError: {},
}

// 判断是否是可重试异常
func IsRetryableError(err error) bool {
	for e, _ := range retryableErrors {
		if strings.Contains(err.Error(), e) {
			return true
		}
	}
	return false
}

// 输出重试错误
func PrintRetryError(prefix string, err error, seconds int64) {
	mylog.Warn(fmt.Sprintf("%v：%v，%d 秒后重试", prefix, err.Error(), seconds))
	time.Sleep(time.Second * time.Duration(seconds))
}
