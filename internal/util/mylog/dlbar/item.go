package dlbar

import (
	"fmt"
	"math"
	"strings"
	"video-downloader-go/internal/util/mylog/color"
)

// Item 是一个日志项的通用接口
type Item interface {
	String(bar *Bar) string
}

const (
	HintItemDownloadBarStart      = "["  // 进度条起始字符
	HintItemDownloadBarEnd        = "]"  // 进度条结尾字符
	HintItemDownloadBarUnfinished = "-"  // 进度条未完成字符
	HintItemDownloadBarFinished   = "*"  // 进度条已完成字符
	HintItemDownloadBarSize       = 12   // 进度条长度
	HintItemSizeB                 = "B"  // 文件大小: B
	HintItemSizeKb                = "KB" // 文件大小: KB
	HintItemSizeMb                = "MB" // 文件大小: MB
	HintItemSizeGb                = "GB" // 文件大小: GB
)

// HintItem 用于在日志中显示当前的状态提示信息
type HintItem struct{}

// NewHintItem 创建一个默认的 HintItem
func NewHintItem() *HintItem { return new(HintItem) }

// String 实现 Item 接口
func (hi *HintItem) String(bar *Bar) string {
	if bar == nil {
		return ""
	}

	// 1 成功和失败状态, 返回对应的提示信息
	if bar.Status == BarStatusOk {
		return color.ToGreen(bar.Hint)
	}
	if bar.Status == BarStatusError {
		return color.ToRed(bar.Hint)
	}

	// 2 正在执行状态, 如果不是 正在下载 子状态, 也是直接返回对应的提示信息
	if bar.ChildStatus != BarChildStatusDownload {
		return color.ToPurple(bar.Hint)
	}

	// 3 正在执行状态, 并且是 正在下载 子状态, 输出进度条信息
	downloadBar := hi.DownloadBar(bar.Percent)
	percent := fmt.Sprintf("%d%%", bar.Percent)
	size := hi.Size(bar.Size)
	return color.ToPurple(fmt.Sprintf("%s %s %s", downloadBar, percent, size))
}

// Size 生成当前的文件大小
func (hi *HintItem) Size(size int64) string {
	fSize := float64(size)
	var interval float64 = 1024
	if fSize < interval {
		return fmt.Sprintf("%.2f%s", fSize, HintItemSizeB)
	}
	fSize /= interval
	if fSize < interval {
		return fmt.Sprintf("%.2f%s", fSize, HintItemSizeKb)
	}
	fSize /= interval
	if fSize < interval {
		return fmt.Sprintf("%.2f%s", fSize, HintItemSizeMb)
	}
	return fmt.Sprintf("%.2f%s", fSize/interval, HintItemSizeGb)
}

// DownloadBar 生成进度条
func (hi *HintItem) DownloadBar(percent int) string {
	tot := HintItemDownloadBarSize - 2
	finish := int(math.Round(float64(tot) * float64(percent) / 100))
	return fmt.Sprintf(
		"[%s%s]",
		strings.Repeat(HintItemDownloadBarFinished, finish),
		strings.Repeat(HintItemDownloadBarUnfinished, tot-finish))
}

var (
	StatusItemExecutingFlags = []string{`\`, `|`, `/`, `-`} // 正在执行的状态标志数组
	StatusItemOk             = "\u2714"                     // 对勾
	StatusItemError          = "\u2718"                     // 错误叉叉
)

// StatusItem 用于在日志末尾显示当前的执行状态
type StatusItem struct {
	// 正在执行的状态的下标
	ExecutingFlagIdx int
}

// NewStatusItem 创建一个默认的 StatusItem
func NewStatusItem() *StatusItem { return new(StatusItem) }

// String 实现 Item 接口
func (si *StatusItem) String(bar *Bar) string {
	if bar == nil {
		return ""
	}

	// 1 成功和失败状态, 直接处理
	if bar.Status == BarStatusOk {
		return color.ToGreen(StatusItemOk)
	}
	if bar.Status == BarStatusError {
		return color.ToRed(StatusItemError)
	}

	// 2 其余状态都认为是正在执行, 循环遍历正在执行的标志进行返回
	cur := StatusItemExecutingFlags[si.ExecutingFlagIdx]
	si.ExecutingFlagIdx = (si.ExecutingFlagIdx + 1) % len(StatusItemExecutingFlags)
	return color.ToPurple(cur)
}

const (
	NameItemMinLen         = len(NameItemOversizeSuffix) + 1 // NameItem 最小长度
	NameItemOversizeSuffix = "..."                           // 超出长度后的后缀
)

// NameItem 用于展示文件名
type NameItem struct {
	Len int
}

// NewNameItem 创建一个默认的 NameItem
func NewNameItem() *NameItem { return NewNameItemWithLen(50) }

// NewNameItemWithLen 创建一个指定长度的 NameItem
func NewNameItemWithLen(l int) *NameItem {
	if l < NameItemMinLen {
		panic(fmt.Sprintf("NameItem 最小长度是 %d", NameItemMinLen))
	}
	return &NameItem{Len: l}
}

// String 实现 Item 接口
func (ni *NameItem) String(bar *Bar) string {
	if bar == nil {
		return ""
	}
	name := bar.Name
	nameRunes := []rune(name)

	// 1 name 长度没有超出 Len 长度, 直接在左边补充相应数量的空格
	sub := ni.Len - len(nameRunes)
	if sub >= 0 {
		return name + strings.Repeat(" ", sub)
	}

	// 2 name 长度超出了 Len 长度, 需要进行截断
	cutStr := string(nameRunes[:ni.Len-len(NameItemOversizeSuffix)])
	return cutStr + NameItemOversizeSuffix
}
