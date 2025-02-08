package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/decoder"
	"video-downloader-go/internal/downloader"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mylog/color"
	"video-downloader-go/internal/util/mylog/dlbar"

	"github.com/pkg/errors"
)

const CurrentVersion = "1.8.5"

func main() {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()

	printBanner()
	fmt.Println("Current Version: ", CurrentVersion)

	// 读取配置
	fmt.Println(color.ToBlue("正在读取配置..."))
	err := config.Load("")
	if err != nil {
		fmt.Println(color.ToRed(err.Error()))
		return
	}
	fmt.Println(color.ToGreen("读取完成"))

	// 读取要处理的视频数据
	fmt.Println(color.ToBlue("正在读取待处理任务..."))
	decodeList, err := readVideoData()
	if err != nil {
		fmt.Println(color.ToRed(err.Error()))
		return
	}
	fmt.Println(color.ToGreen("读取完成"))
	if decodeList.Empty() {
		fmt.Println(color.ToBlue("解析列表为空，程序停止"))
		return
	}
	downloadList := new(meta.TaskDeque[meta.Download])
	fmt.Println(color.ToGreen("程序初始化完成, 开始处理任务..."))

	mylog.Start()

	// 开启解析任务
	decoder.ListenAndDecode(decodeList, func(d *meta.Download) {
		downloadList.OfferLast(d)
	})

	// 开启下载任务
	var downloadWg sync.WaitGroup
	remainCnt := int64(decodeList.Size())
	downloadWg.Add(decodeList.Size())
	downloader.ListenAndDownload(downloadList, func() {
		downloadWg.Done()
		atomic.AddInt64(&remainCnt, -1)
		mylog.Successf("一个文件下载完成，剩余：%v 个", remainCnt)
	}, func(dmt *meta.Download) {
		// 下载器判断出无法正常下载的视频，重新加入到解析列表中
		fileName, originUrl := dmt.FileName, dmt.OriginUrl
		dmt.LogBar.WaitingHint("正在等待解析")
		decodeList.OfferLast(&meta.Video{Name: fileName, Url: originUrl, LogBar: dmt.LogBar})
	})
	downloadWg.Wait()
	mylog.Success("所有任务处理完成")
}

// printBanner 输出 banner
func printBanner() {
	bannerBytes, err := os.ReadFile("config/banner.txt")
	if err != nil {
		return
	}
	fmt.Println(string(bannerBytes))
}

// readVideoData 读取用户在 config/data.txt 目录下配置的输入数据
func readVideoData() (*meta.TaskDeque[meta.Video], error) {
	mylog.Info("正在读取源数据文件 data.txt...")

	// 打开文件
	f, err := os.Open("config/data.txt")
	if err != nil {
		return nil, errors.Wrap(err, "打开源数据文件失败")
	}
	defer f.Close()

	// 初始化读取器
	scanner := bufio.NewScanner(f)

	// 逐行读取数据并处理
	list := new(meta.TaskDeque[meta.Video])
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// 忽略空行
			continue
		}
		arr := strings.Split(line, "|")
		if len(arr) != 2 {
			return nil, errors.New("文件格式不合法，请遵循：`文件名|地址`")
		}
		bar := dlbar.NewBar(
			dlbar.WithStatus(dlbar.BarStatusWaiting),
			dlbar.WithHint("正在等待解析"),
			dlbar.WithName(arr[0]),
		)
		list.OfferLast(&meta.Video{LogBar: bar, Name: arr[0], Url: arr[1]})
		mylog.GlobalPanel.RegisterBar(bar)
	}

	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "扫描源数据文件失败")
	}

	list.Range(func(item *meta.Video, index int) {
		mylog.Infof("%v", item)
	})

	mylog.Success("读取完成！")
	return list, nil
}
