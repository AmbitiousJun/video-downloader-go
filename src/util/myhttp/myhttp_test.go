package myhttp_test

import (
	"bufio"
	"fmt"
	"net/http"
	"testing"
	"video-downloader-go/src/util/myhttp"
)

func TestGetRequestRanges(t *testing.T) {
	headers := map[string]string{
		"Range": "bytes=3-333",
	}
	ranges, err := myhttp.GetRequestRanges("https://blog.ambitiousjun.cn", "GET", headers)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(ranges)
}

func TestTimeoutHttpClient(t *testing.T) {
	client := myhttp.TimeoutHttpClient()
	req, err := http.NewRequest("GET", "https://google.com", nil)
	if err != nil {
		t.Error(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
	if err = scanner.Err(); err != nil {
		t.Error(err)
	}
}
