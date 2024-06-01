package mystring

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// PadRightByRuneWidth 根据字符串的显示宽度, 补充相应数量的占位符
func PadRightByRuneWidth(s string, maxLen int, placeholder byte) string {
	width := runewidth.StringWidth(s)
	if width > maxLen {
		return s
	}
	return s + strings.Repeat(string(placeholder), maxLen-width)
}
