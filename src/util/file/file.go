package file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"video-downloader-go/src/util/log"

	"github.com/pkg/errors"
)

// 删除目录
// @param dirPath 要删除目录的绝对路径
// @return 是否删除成功
func DeleteDir(dirPath string) bool {
	fileInfo, err := os.Stat(dirPath)
	if err != nil || !fileInfo.IsDir() {
		// 没有这个目录，或者不是目录
		return false
	}
	err = os.RemoveAll(dirPath)
	return err == nil
}

// 生成一个用于下载 ts 文件的临时目录
// @param filename 文件名称
// @return 生成的临时目录绝对路径
func InitTempTsDir(filename string) (string, error) {
	suffix := "ts_dir"
	// TODO: suffix 暂时写死，需要从配置中读取
	dirPath := fmt.Sprintf("%v_%v", filename, suffix)
	_, err := os.Stat(dirPath)
	if err == nil {
		log.Warn("临时目录已存在：" + dirPath)
		return dirPath, nil
	}
	err = os.MkdirAll(dirPath, fs.ModeDir)
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
