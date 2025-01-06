//go:build windows
// +build windows

package config

import (
	"fmt"
	"video-downloader-go/internal/util/mylog/color"
)

// increaseSystemUlimit 增大系统最多可打开的文件描述符个数
func increaseSystemUlimit(limit uint64) {
	fmt.Println(color.ToGray("windows 系统暂不支持修改文件描述符最大个数"))
}
