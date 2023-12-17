package ytdl_test

import (
	"os"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/decoder/ytdl"
	"video-downloader-go/internal/util/mylog"
)

// 测试 format code 解析
func TestCodeSelector(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	config.Load("../../../config/config.yml")

	// 模拟控制台输入
	originStdIn := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() {
		os.Stdin = originStdIn
		w.Close()
	}()
	w.Write([]byte("\n"))
	w.Write([]byte("1234\n"))

	url := "https://www.mgtv.com/b/600010/20245077.html?fpa=1217&fpos=&lastp=ch_home"
	slt := &ytdl.CodeSelector{Url: url}
	if code, err := slt.RequestCode(); err != nil {
		t.Error(err)
	} else {
		mylog.Successf("解析成功：%v", code)
	}
}
