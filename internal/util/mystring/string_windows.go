//go:build windows
// +build windows

package mystring

import (
	"golang.org/x/text/encoding/simplifiedchinese"
)

// UTF8 将一个字符串的字符编码转换成 utf8 后返回
// 在 Windows 系统中，经常出现 GBK 编码，需要特殊处理
func UTF8(raw string) string {
	output, err := simplifiedchinese.GBK.NewDecoder().Bytes([]byte(raw))
	if err != nil {
		return raw
	}
	return string(output)
}
