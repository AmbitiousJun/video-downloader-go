package ffmpeg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"video-downloader-go/internal/constant"
	"video-downloader-go/internal/util/https"
	"video-downloader-go/internal/util/mylog/color"
)

const (

	// ReleasePage ffmpeg 发布页地址
	ReleasePage = "https://github.com/AmbitiousJun/ffmpeg-release/releases/latest/download"
)

// arch2ExecNameMap 根据系统的芯片架构, 映射到对应的二进制文件
var arch2ExecNameMap = map[string]string{
	"darwin/amd64":  "ffmpeg_macos",
	"darwin/arm64":  "ffmpeg_macos",
	"windows/386":   "ffmpeg.exe",
	"windows/amd64": "ffmpeg.exe",
	"windows/arm":   "ffmpeg.exe",
	"windows/arm64": "ffmpeg.exe",
	"linux/386":     "ffmpeg_linux_386",
	"linux/amd64":   "ffmpeg_linux_amd64",
	"linux/arm":     "ffmpeg_linux_arm",
	"linux/arm64":   "ffmpeg_linux_arm64",
}

var (
	parentPath = filepath.Join(constant.Dir_DataRoot, "ffmpeg") // 二进制文件存放根路径
	execPath   string                                           // 根据当前系统架构自动生成一个二进制文件地址
	execOk     bool                                             // 标记二进制是否检测通过
)

func ExecPath() string {
	return execPath
}

// AutoDownloadExec 自动根据系统架构下载对应版本的 ffmpeg 到数据目录下
//
// 下载失败只会进行日志输出, 不会影响到程序运行
func AutoDownloadExec() error {
	// 获取系统架构
	gos, garch := runtime.GOOS, runtime.GOARCH

	// 生成二进制文件地址
	execName, ok := arch2ExecNameMap[fmt.Sprintf("%s/%s", gos, garch)]
	if !ok {
		return fmt.Errorf("不支持的芯片架构: %s/%s, ffmpeg 相关功能失效", gos, garch)
	}
	execPath = fmt.Sprintf("%s/%s", parentPath, execName)

	defer func() {
		if execOk {
			execPath, _ = filepath.Abs(execPath)
		}
	}()

	// 如果文件不存在, 触发自动下载
	stat, err := os.Stat(execPath)
	if err == nil {
		if stat.IsDir() {
			return fmt.Errorf("二进制文件路径被目录占用: %s, 请手动处理后尝试重启服务", execPath)
		}
		execOk = true
		fmt.Println(color.ToGreen("ffmpeg 环境检测通过 ✓"))
		return nil
	}

	fmt.Println(color.ToBlue("检测不到 ffmpeg 环境, 即将开始自动下载"))

	if err = os.MkdirAll(parentPath, os.ModePerm); err != nil {
		return fmt.Errorf("数据目录异常: %s, err: %v", parentPath, err)
	}

	fmt.Printf(color.ToBlue("ffmpeg 下载发布页: %s\n"), ReleasePage)

	_, resp, err := https.Request(http.MethodGet, ReleasePage+"/"+execName, nil, nil, true)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if !https.IsSuccessCode(resp.StatusCode) {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	execFile, err := os.OpenFile(execPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("初始化二进制文件路径失败: %s, err: %v", execPath, err)
	}
	defer execFile.Close()
	io.Copy(execFile, resp.Body)

	// 标记就绪状态
	fmt.Printf(color.ToGreen("ffmpeg 自动下载成功 ✓, 路径: %s\n"), execPath)
	execOk = true
	return nil
}
