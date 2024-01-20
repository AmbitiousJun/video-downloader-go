package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	WindowsFileSuffix = ".exe"

	DistDir   = "../../dist"                 // 存放打包完成的可执行文件路径
	ConfigDir = "../../config"               // 配置文件目录
	OutputDir = "/Users/ambitious/Downloads" // zip 输出目录
	// OutputDir = "C:/Users/Ambitious/Downloads" // zip 输出目录
	ExecName = "start" // 可执行文件名称
)

// 需要打包的平台
var Platforms = []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64", "windows-amd64", "windows-arm64"}

// 存放错误信息的通道
var ErrChan = make(chan error)

func main() {
	// 遍历 dist 目录
	dst, err := os.Stat(DistDir)
	if err != nil || !dst.IsDir() {
		panic("dist 目录检测失败")
	}

	var wg sync.WaitGroup
	var totalTasks atomic.Int64

	go func() {
		err = <-ErrChan
		close(ErrChan)
		// 结束掉仍在执行的 goroutine
		wg.Add(-int(totalTasks.Load()))
	}()

	err = filepath.WalkDir(DistDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == DistDir {
			return err
		}
		if d.IsDir() {
			// 跳过子目录
			return filepath.SkipDir
		}
		fileName := strings.TrimSuffix(d.Name(), WindowsFileSuffix)
		wg.Add(1)
		totalTasks.Add(1)
		go func() {
			defer func() {
				wg.Done()
				totalTasks.Add(-1)
			}()
			createZip(OutputDir+"/"+fileName+".zip", path)
		}()
		return err
	})

	wg.Wait()

	if err != nil {
		panic(errors.Wrap(err, "打包失败"))
	}

	log.Println("打包完成！")
}

var errOnce sync.Once

func createZip(zipPath, execPath string) {
	var err error
	defer func() {
		if err != nil {
			errOnce.Do(func() {
				ErrChan <- err
			})
		}
	}()

	// 创建一个新的 zip 文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return
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
		return
	}

	// 添加配置目录
	err = addDirToZip(zipWriter, ConfigDir, "config")
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
		if d.IsDir() {
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
