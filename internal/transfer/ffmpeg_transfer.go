package transfer

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/myfile"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mylog/dlbar"
	"video-downloader-go/internal/util/mymath"

	"github.com/pkg/errors"
)

// concatFileFunc 合并文件函数
type concatFileFunc func(tsDir string, tsFilePaths []string, outputPath string, bar *dlbar.Bar) error

type ffmpegTransfer struct {
	concatFileFunc concatFileFunc
}

func (ft *ffmpegTransfer) Ts2Mp4(tsDir, outputPath string, bar *dlbar.Bar) error {
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
	regex, err := regexp.Compile(config.G.Transfer.TsFilenameRegex)
	if err != nil {
		return errors.Wrap(err, "正则表达式编译错误")
	}
	sort.Slice(tsFilePaths, func(i, j int) bool {
		bi, bj := filepath.Base(tsFilePaths[i]), filepath.Base(tsFilePaths[j])
		mi, mj := regex.FindStringSubmatch(bi), regex.FindStringSubmatch(bj)
		if len(mi) == 0 || len(mj) == 0 {
			err = errors.New("ts 文件名不规范")
			return false
		}
		in, _ := strconv.Atoi(mi[1])
		jn, _ := strconv.Atoi(mj[1])
		return in < jn
	})
	if err != nil {
		return errors.Wrap(err, "排序文件失败")
	}
	err = ft.concatFileFunc(tsDir, tsFilePaths, outputPath, bar)
	if err != nil {
		return errors.Wrap(err, "合并 ts 文件时出现错误")
	}
	return nil
}

// ConcatFilesByTxt 先将 ts 切片编排到 txt 文件中, 再调用 ffmpeg 一次性合并
func ConcatFilesByTxt(tsDir string, tsFilePaths []string, outputPath string, bar *dlbar.Bar) error {
	bar.TransferHint("正在合并切片文件 (0%)")

	// 1 将切片信息写入 tsDir
	filelistContent := strings.Builder{}
	for idx, tsPath := range tsFilePaths {
		baseName := filepath.Base(tsPath)
		filelistContent.WriteString(fmt.Sprintf("file '%s'", baseName))
		if idx < len(tsFilePaths)-1 {
			filelistContent.WriteByte('\n')
		}
	}
	filelistPath := filepath.Join(tsDir, "filelist.txt")
	if err := os.WriteFile(filelistPath, []byte(filelistContent.String()), os.ModePerm); err != nil {
		return fmt.Errorf("写入切片编排信息失败: %v", err)
	}
	bar.TransferHint("正在合并切片文件 (50%)")

	// 2 调用 ffmpeg 进行合并
	cmd := exec.Command(config.FfmpegPath, "-f", "concat", "-safe", "0", "-i", filelistPath, "-c", "copy", outputPath)
	if err := executeCmd(cmd); err != nil {
		return fmt.Errorf("调用 ffmpeg 出现异常: %v", err)
	}
	bar.TransferHint("正在合并切片文件 (100%)")
	return nil
}

// 核心的合并 ts 文件逻辑
func ConcatFilesByStr(tsDir string, tsFilePaths []string, outputPath string, bar *dlbar.Bar) error {
	tempTsFilePath := fmt.Sprintf("%s/ts_%d.ts", tsDir, math.MaxInt32)
	tempDestFilePath := strings.Replace(outputPath, ".mp4", ".ts", -1)
	if e, d := myfile.DeleteFileIfExist(tempTsFilePath); e && !d {
		return errors.New("无法删除临时文件：" + tempTsFilePath)
	}
	if e, d := myfile.DeleteFileIfExist(tempDestFilePath); e && !d {
		return errors.New("无法删除临时文件：" + tempDestFilePath)
	}
	// 遍历列表合成
	size, current := len(tsFilePaths), 0
	for current < size {
		// 一次性合并 50 个分片
		handleSize := int(mymath.Min(50, int64(size-current)))
		concatBuilder := &strings.Builder{}
		concatBuilder.WriteString("concat:")
		if myfile.FileExist(tempTsFilePath) {
			concatBuilder.WriteString(tempTsFilePath)
		}
		for i := 0; i < handleSize; i++ {
			pos := current + i
			if tsFilePaths[pos] == tempTsFilePath {
				// 不处理临时 ts 文件
				continue
			}
			if i != 0 || myfile.FileExist(tempTsFilePath) {
				// 非首次合并，需要 |
				concatBuilder.WriteString("|")
			}
			concatBuilder.WriteString(tsFilePaths[pos])
		}
		current += handleSize
		concat := concatBuilder.String()
		cmd := exec.Command(config.FfmpegPath, "-i", concat, "-c", "copy", tempDestFilePath)
		err := executeCmd(cmd)
		if err != nil {
			return errors.Wrap(err, "执行 ffmpeg 合并命令失败")
		}
		if e, d := myfile.DeleteFileIfExist(tempTsFilePath); e && !d {
			return errors.New("无法删除临时文件：" + tempTsFilePath)
		}
		if myfile.FileExist(tempDestFilePath) {
			if err = os.Rename(tempDestFilePath, tempTsFilePath); err != nil {
				return errors.Wrap(err, "临时文件拷贝异常："+tempDestFilePath)
			}
		}
		percent := int(math.Round(float64(current) / float64(size) * 100))
		bar.TransferHint(fmt.Sprintf("正在合并分片(%d%%)", percent))
	}
	// 全部转换完成后，生成最终文件
	if !myfile.FileExist(tempTsFilePath) {
		return errors.New("检测不到最终的 ts 文件")
	}
	cmd := exec.Command(config.FfmpegPath, "-i", "concat:"+tempTsFilePath, "-c", "copy", outputPath)
	if err := executeCmd(cmd); err != nil {
		return errors.Wrap(err, "合并最终视频文件失败")
	}
	if e, d := myfile.DeleteFileIfExist(tempTsFilePath); e && !d {
		mylog.Warn("临时 ts 删除失败")
	}
	return nil
}

// 执行命令行命令
func executeCmd(cmd *exec.Cmd) error {
	_, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "执行命令时出错")
	}
	return nil
}
