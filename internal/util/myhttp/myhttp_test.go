package myhttp_test

import (
	"bufio"
	"fmt"
	"net/http"
	"testing"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/util/myhttp"
)

func TestDownloadWithRateLimit(t *testing.T) {
	err := config.Load("D:\\学习\\Go\\projects\\video-downloader-go\\config\\config.yml")
	if err != nil {
		t.Error(err)
	}
	u := "https://xy180x105x103x113xy.mcdn.bilivideo.cn:8082/v1/resource/1336749181-1-100110.m4s?agrr=1&build=0&buvid=C26216B6-DA78-6124-413A-239D03A36AC766301infoc&bvc=vod&bw=26787&deadline=1700739669&e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M%3D&f=u_0_0&gen=playurlv2&logo=A0000001&mcdnid=1002547&mid=0&nbs=1&nettype=0&oi=1946640600&orderid=0%2C3&os=mcdn&platform=pc&sign=b837a7&traceid=trkPyouBwIKzOt_0_e_N&uipk=5&uparams=e%2Cuipk%2Cnbs%2Cdeadline%2Cgen%2Cos%2Coi%2Ctrid%2Cmid%2Cplatform&upsig=0ec9184610875a3d3af0c823a83fc0b6"
	req, err := http.NewRequest("GET", u, nil)
	header := myhttp.GenDefaultHeaderMapByUrl(nil, u)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if err != nil {
		t.Error(err)
	}
	err = myhttp.DownloadWithRateLimit(req, "C:/Users/Ambitious/Downloads/1.mp4")
	if err != nil {
		t.Error(err)
	}
}

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
