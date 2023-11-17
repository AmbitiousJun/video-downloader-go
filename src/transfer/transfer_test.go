package transfer_test

import (
	"testing"
	"video-downloader-go/src/transfer"
)

func TestFfmpegTransfer(t *testing.T) {
	transfer := transfer.NewFfmpegTransfer()
	err := transfer.Ts2Mp4("C:/Users/Ambitious/Downloads/1_ts_dir", "")
	if err != nil {
		t.Error(err)
	}
}
