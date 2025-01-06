//go:build !windows
// +build !windows

package config

import (
	"fmt"
	"syscall"
	"video-downloader-go/internal/util/mylog/color"
)

// increaseSystemUlimit 增大系统最多可打开的文件描述符个数
func increaseSystemUlimit(limit uint64) {
	var rLimit syscall.Rlimit

	// 获取当前文件描述符限制
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Printf(color.ToRed("修改文件描述符最大个数失败: %v"), err)
		return
	}

	// 更新文件描述符限制
	rLimit.Cur = limit
	if limit > rLimit.Max {
		rLimit.Max = limit
	}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Printf(color.ToRed("修改文件描述符最大个数失败: %v"), err)
	}
}
