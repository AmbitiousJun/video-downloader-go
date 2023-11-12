package file

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// 初始化文件的父目录
// @path 文件的绝对路径
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
