package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	WindowsFileSuffix = ".exe"

	DistDir   = "../../dist"                   // 存放打包完成的可执行文件路径
	ConfigDir = "../../config"                 // 配置文件目录
	OutputDir = "C:/Users/Ambitious/Downloads" // zip 输出目录
	ExecName  = "start"                        // 可执行文件名称
)

// 需要打包的平台
var Platforms = []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64", "windows-amd64", "windows-arm64"}

func main() {
	// 遍历 dist 目录
	dst, err := os.Stat(DistDir)
	if err != nil || !dst.IsDir() {
		panic("dist 目录检测失败")
	}
	err = filepath.WalkDir(DistDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return err
		}
		fileName := strings.TrimSuffix(d.Name(), WindowsFileSuffix)
		return createZip(OutputDir+"/"+fileName+".zip", path)
	})

	if err != nil {
		panic(errors.Wrap(err, "打包失败"))
	}

	log.Println("打包完成！")
}

func createZip(zipPath, execPath string) error {
	// 创建一个新的 zip 文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建一个 zip writter
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加可执行文件
	execName := ExecName
	if strings.HasSuffix(execPath, WindowsFileSuffix) {
		execName += WindowsFileSuffix
	}
	if err = addFileToZip(zipWriter, execPath, execName); err != nil {
		return err
	}

	// 添加配置目录
	if err = addDirToZip(zipWriter, ConfigDir, "config"); err != nil {
		return err
	}
	return nil
}

func addDirToZip(zipWriter *zip.Writer, fileName, zipName string) error {
	// 遍历目录中的文件
	return filepath.WalkDir(fileName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == fileName {
			// 不处理根目录
			return nil
		}
		if d.IsDir() && path != fileName {
			// 递归添加子目录
			err = addDirToZip(zipWriter, path, zipName+"/"+filepath.Base(path))
			if err == nil {
				err = filepath.SkipDir
			}
			return err
		}
		// 添加文件
		return addFileToZip(zipWriter, path, zipName+"/"+filepath.Base(path))
	})
}

func addFileToZip(zipWriter *zip.Writer, fileName, zipName string) error {
	log.Printf("正在打包文件【%v】，目标【%v】\n", fileName, zipName)
	// 打开要添加到 zip 的文件
	fileToZip, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// 创建一个 writer
	writer, err := zipWriter.Create(zipName)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
