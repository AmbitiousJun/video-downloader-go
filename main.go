package main

import (
	"bufio"
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

	"github.com/pkg/errors"
)

func main() {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()

	// 读取配置
	err := config.Load("")
	if err != nil {
		mylog.Error(err.Error())
		return
	}

	// 读取要处理的视频数据
	decodeList, err := readVideoData()
	if err != nil {
		mylog.Error(err.Error())
		return
	}
	if decodeList.Empty() {
		mylog.Info("解析列表为空，程序停止")
		return
	}
	downloadList := new(meta.TaskDeque[meta.Download])

	// 输出初始化日志
	mylog.PrintAllLogs()

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
		decodeList.OfferLast(&meta.Video{Name: fileName, Url: originUrl})
	})
	downloadWg.Wait()
	mylog.Success("所有任务处理完成")
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
		list.OfferLast(&meta.Video{Name: arr[0], Url: arr[1]})
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
