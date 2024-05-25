package dlbar_test

import (
	"fmt"
	"testing"
	"video-downloader-go/internal/util/mylog/dlbar"
)

func TestBar(t *testing.T) {
	b := dlbar.NewBar(
		dlbar.WithPercent(75),
		dlbar.WithSize(1610612736),
		dlbar.WithStatus(dlbar.BarStatusExecuting),
		dlbar.WithChildStatus(dlbar.BarChildStatusDownload),
		dlbar.WithName("The.Truth.S02E08.2024-05-24.第4期下：《走不出的忙活街》-抓马认亲"),
	)
	fmt.Println(b.String())
}

func TestHintItem(t *testing.T) {
	item := dlbar.NewHintItem()
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusError, Hint: "下载失败"}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusOk, Hint: "下载完成"}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusDecode, Hint: "正在解析..."}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusTransfer, Hint: "正在转换(1/50)..."}))
	fmt.Println(item.String(&dlbar.Bar{Percent: 0, Size: 0, Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusDownload}))
	fmt.Println(item.String(&dlbar.Bar{Percent: 25, Size: 536870912, Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusDownload}))
	fmt.Println(item.String(&dlbar.Bar{Percent: 50, Size: 1073741824, Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusDownload}))
	fmt.Println(item.String(&dlbar.Bar{Percent: 75, Size: 1610612736, Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusDownload}))
	fmt.Println(item.String(&dlbar.Bar{Percent: 100, Size: 2147483648, Status: dlbar.BarStatusExecuting, ChildStatus: dlbar.BarChildStatusDownload}))
}

func TestNameItem(t *testing.T) {
	item := dlbar.NewNameItem()
	fmt.Println(item.String(&dlbar.Bar{Name: "The.Truth.S02E08.2024-05-24.第4期下：《走不出的忙活街》-抓马认亲"}))
	fmt.Println(item.String(&dlbar.Bar{Name: "The.Truth.S02E09.2024-05-24"}))
}

func TestStatusItem(t *testing.T) {
	item := dlbar.NewStatusItem()
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusOk}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusError}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
	fmt.Println(item.String(&dlbar.Bar{Status: dlbar.BarStatusExecuting}))
}
