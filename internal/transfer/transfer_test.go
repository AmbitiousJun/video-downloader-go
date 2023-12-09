package transfer_test

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"testing"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/transfer"
)

// 测试逐行输出命令行结果
func TestCmd(t *testing.T) {
	cmd := exec.Command("ffmpeg", "--help")
	var err error
	var out io.ReadCloser
	if out, err = cmd.StdoutPipe(); err != nil {
		t.Error(err)
		return
	}
	if err = cmd.Start(); err != nil {
		t.Error(err)
		return
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err = cmd.Wait(); err != nil {
		t.Error(err)
	}
}

func TestFfmpegTransfer(t *testing.T) {
	err := config.Load("../../config/config.yml")
	if err != nil {
		t.Error(err)
		return
	}
	ft := transfer.Instance()
	err = ft.Ts2Mp4("/Users/ambitious/Downloads/测试.mp4_temp_ts_files", "/Users/ambitious/Downloads/测试.mp4")
	if err != nil {
		t.Error(err)
	}
}
