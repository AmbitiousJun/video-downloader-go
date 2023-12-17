package myfile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"video-downloader-go/internal/util/mylog"

	"github.com/pkg/errors"
)

// 检查文件是否存在
func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// DeleteAnyFileContainsPrefix 将识别给定的绝对路径 fp 所在的目录路径下
// 任何已 fp 文件名为前缀的 目录 或者 文件
// 返回值分别是识别到的文件数和成功删除的文件数
func DeleteAnyFileContainsPrefix(fp string) (int, int, error) {
	scanCnt, delCnt := 0, 0
	// 1 分离出目录路径和文件名
	dirName, fileName := filepath.Dir(fp), filepath.Base(fp)

	// 2 遍历目录下的所有文件，进行筛选和删除
	pList := []string{}
	filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		// 判断是否包含前缀
		if strings.HasPrefix(filepath.Base(path), fileName) {
			pList = append(pList, path)
		}
		return nil
	})

	// 遍历删除
	for _, path := range pList {
		var err error
		scanCnt++
		if s, _ := os.Stat(path); s.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
		if err != nil {
			return -1, -1, err
		}
		delCnt++
	}

	return scanCnt, delCnt, nil
}

// 删除文件，如果存在
// @return 第一个参数表示文件是否存在，第二个参数表示删除是否成功
func DeleteFileIfExist(filePath string) (bool, bool) {
	if !FileExist(filePath) {
		return false, false
	}
	err := os.Remove(filePath)
	if err != nil {
		return true, false
	}
	return true, true
}

// 删除目录
// @param dirPath 要删除目录的绝对路径
// @return 是否删除成功
func DeleteDir(dirPath string) error {
	fileInfo, err := os.Stat(dirPath)
	if err != nil || !fileInfo.IsDir() {
		// 没有这个目录，或者不是目录
		return errors.New("目录不存在或不是目录")
	}
	return os.RemoveAll(dirPath)
}

// 生成一个用于下载 ts 文件的临时目录
// @param filename 文件名称
// @param suffix 临时目录后缀
// @return 生成的临时目录绝对路径
func InitTempTsDir(filename, suffix string) (string, error) {
	dirPath := fmt.Sprintf("%v_%v", filename, suffix)
	_, err := os.Stat(dirPath)
	if err == nil {
		mylog.Warn("临时目录已存在：" + dirPath)
		return dirPath, nil
	}
	err = os.MkdirAll(dirPath, fs.ModePerm|fs.ModeDir)
	if err != nil {
		return "", errors.Wrap(err, "创建临时目录失败")
	}
	return dirPath, nil
}

// 初始化文件的父目录
// @param path 文件的绝对路径
func InitFileDirs(path string) error {
	// 1 获取文件的父目录的绝对路径
	parentPath, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return errors.Wrapf(err, "无法获获取文件的父目录，path：%v", path)
	}
	// 2 创建目录
	_, err = os.Stat(parentPath)
	if err != nil {
		// 目录不存在，需要创建
		err = os.MkdirAll(parentPath, 0755)
		if err != nil {
			return errors.Wrapf(err, "创建父目录失败，path：%v", parentPath)
		}
	}
	// 3 如果文件存在，将其删除
	_, err = os.Stat(path)
	if err == nil {
		err = os.Remove(path)
		if err != nil {
			return errors.Wrapf(err, "文件删除失败，path：%v", path)
		}
	}
	return nil
}
