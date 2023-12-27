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
	w.Write([]byte("928\n"))
	// w.Write([]byte("248+251\n"))
	// w.Write([]byte("100050+30280\n"))

	url := "https://www.mgtv.com/b/593651/20291328.html"
	// url = "https://www.youtube.com/watch?v=OfIFA-V6Zec"
	// url = "https://www.bilibili.com/video/BV18e411B7HF"
	slt := ytdl.NewCodeSelector(url)

	for i := 0; i < 2; i++ {
		if code, err := slt.RequestCode(); err != nil {
			t.Error(err)
		} else {
			mylog.Successf("解析成功：%v", code)
		}
	}
}
