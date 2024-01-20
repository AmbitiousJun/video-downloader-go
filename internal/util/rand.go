package util

import (
	"math/rand"
	"time"
)

const (
	LetterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

// RandString 生成一个随机字符串
//
// 当 n > 0 时，返回指定长度的随机字符串
//
// 当 n <= 0 时，返回空字符串
func RandString(n int) string {
	if n <= 0 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = LetterBytes[r.Intn(len(LetterBytes))]
	}

	return string(b)
}
