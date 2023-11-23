// 提供一些重试工具函数
package util

import (
	"fmt"
	"time"
	"video-downloader-go/internal/util/log"
)

// 输出重试错误
func PrintRetryError(prefix string, err error, seconds time.Duration) {
	log.Warn(fmt.Sprintf("%v：%v，%v 秒后重试", prefix, err.Error(), seconds))
	time.Sleep(time.Second * seconds)
}
