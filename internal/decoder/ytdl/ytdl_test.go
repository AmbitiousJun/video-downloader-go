package ytdl_test

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/decoder/ytdl"
	"video-downloader-go/internal/util/mystring"
)

// 测试生成 format code
func TestGenFormat(t *testing.T) {
	cmd := exec.Command("youtube-dl", "-F", "https://www.mgtv.com/b/601878/20284878.html?fpa=se&lastp=so_result")
	output, err := cmd.Output()
	if err != nil {
		t.Error(err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(mystring.UTF8(string(output))))
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
	}
}

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
	// w.Write([]byte("928\n"))
	// w.Write([]byte("248+251\n"))
	w.Write([]byte("100050+30280\n"))

	urls := []string{
		"https://www.bilibili.com/video/BV1LQ4y1V79r",
		"https://www.bilibili.com/video/BV1R64y1j7N2",
	}

	for _, url := range urls {
		slt := ytdl.NewCodeSelector(url)
		if code, err := slt.RequestCode(); err != nil {
			t.Error(err)
			return
		} else {
			log.Println("解析成功", code)
		}
	}
}
