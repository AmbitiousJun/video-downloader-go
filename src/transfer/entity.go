package transfer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"video-downloader-go/src/config"

	"github.com/pkg/errors"
)

// ts 文件转换器接口
type TsTransfer interface {
	// 将 ts 格式的文件列表转换成 mp4 格式的视频文件
	// @param tsDir 存放 ts 文件的目录
	// @param outputPath 合并后输出的文件绝对地址
	Ts2Mp4(tsDir, outputPath string) error
}

type ffmpegTransfer struct{}

func NewFfmpegTransfer() TsTransfer {
	return &ffmpegTransfer{}
}

func (ft *ffmpegTransfer) Ts2Mp4(tsDir, outputPath string) error {
	fi, err := os.Stat(tsDir)
	if err != nil || !fi.IsDir() {
		return errors.New("无效的 ts 目录")
	}
	// 1 读取文件并排序
	tsFilePaths := []string{}
	err = filepath.Walk(tsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "遍历目录文件失败")
		}
		if info.IsDir() {
			return nil
		}
		tsFilePaths = append(tsFilePaths, path)
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "读取文件时出错")
	}
	regex := regexp.MustCompile(config.GlobalConfig.Transfer.TsFilenameRegex)
	sort.Slice(tsFilePaths, func(i, j int) bool {
		bi, bj := filepath.Base(tsFilePaths[i]), filepath.Base(tsFilePaths[j])
		mi, mj := regex.FindAllString(bi, -1), regex.FindAllString(bj, -1)
		fmt.Println(mi, mj)
		if len(mi) != 1 || len(mj) != 1 {
			err = errors.New("ts 文件名不规范")
			return false
		}
		in, _ := strconv.Atoi(mi[0])
		jn, _ := strconv.Atoi(mj[0])
		return in < jn
	})
	if err != nil {
		return errors.Wrap(err, "排序文件失败")
	}
	fmt.Println(tsFilePaths)
	return nil
}
