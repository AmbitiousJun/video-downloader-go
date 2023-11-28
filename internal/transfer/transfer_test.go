package transfer_test

import (
	"testing"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/transfer"
)

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
