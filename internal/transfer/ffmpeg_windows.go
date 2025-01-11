//go:build windows
// +build windows

package transfer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/mylog/dlbar"
)

// 核心的合并 ts 文件逻辑
func ConcatFilesByStrV2(tsDir string, tsFilePaths []string, outputPath string, bar *dlbar.Bar) error {
	bar.TransferHint("正在合并切片 (0%)")

	// 1 准备 shell 脚本命令
	shellBuilder := strings.Builder{}
	shellBuilder.WriteString("@echo off\r\n")
	shellBuilder.WriteString("chcp 65001\r\n")
	shellBuilder.WriteString(`set "SCRIPT_DIR=%~dp0"`)
	shellBuilder.WriteString("\r\n")
	shellBuilder.WriteString(`cd "%SCRIPT_DIR%"`)
	shellBuilder.WriteString("\r\n")
	shellBuilder.WriteString(fmt.Sprintf(`set "FFMPEG=%s"`, config.FfmpegPath))
	shellBuilder.WriteString("\r\n")
	shellBuilder.WriteString(fmt.Sprintf(`set "OUTPUT=%s"`, outputPath))
	shellBuilder.WriteString("\r\n")
	bar.TransferHint("正在合并切片 (25%)")

	concatPlaceholder := "{{concat}}"
	ffmpegCmd := fmt.Sprintf(`"%%FFMPEG%%" -i "concat:%s" -c copy "%%OUTPUT%%"`, concatPlaceholder)
	concatBuilder := strings.Builder{}
	for idx, tsPath := range tsFilePaths {
		if idx != 0 {
			concatBuilder.WriteString("|")
		}
		concatBuilder.WriteString(filepath.Base(tsPath))
	}
	ffmpegCmd = strings.Replace(ffmpegCmd, concatPlaceholder, concatBuilder.String(), -1)
	shellBuilder.WriteString(ffmpegCmd)
	shellBuilder.WriteString("\n")
	bar.TransferHint("正在合并切片 (50%)")

	// 2 将命令写入文件 merge.bat 中
	mergeScriptName := filepath.Join(tsDir, "merge.bat")
	if err := os.WriteFile(mergeScriptName, []byte(shellBuilder.String()), os.ModePerm); err != nil {
		return fmt.Errorf("写入 merge 脚本失败: %v, path: %s", err, mergeScriptName)
	}
	defer os.Remove(mergeScriptName)
	bar.TransferHint("正在合并切片 (75%)")

	// 3 执行脚本进行合并
	cmd := exec.Command(mergeScriptName)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行脚本失败: %v, script: %s", err, shellBuilder.String())
	}
	bar.TransferHint("正在合并切片 (100%)")

	return nil
}
