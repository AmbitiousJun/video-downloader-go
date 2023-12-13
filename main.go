package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
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

	// TODO 开启解析任务

	// TODO 开启下载任务
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
	list := &meta.TaskDeque[meta.Video]{}
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

	for i := 0; i < list.Size(); i++ {
		mylog.Success(fmt.Sprintf("%v", list.Get(i)))
	}

	mylog.Success("读取完成！")
	return list, nil
}
