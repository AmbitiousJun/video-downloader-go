package coredl_test

import (
	"fmt"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/log"
)

// 测试下载 MP4
func TestDownloadMp4(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	config.Load("../../../config/config.yml")
	dlpool.InitFromGlobalConfig()
	url := "https://xy123x183x124x190xy.mcdn.bilivideo.cn:8082/v1/resource/1287066422-1-100113.m4s?agrr=1&build=0&buvid=291519AA-71D6-187E-0BA3-C38B5392A4B653789infoc&bvc=vod&bw=66803&deadline=1701428580&e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M%3D&f=u_0_0&gen=playurlv2&logo=A0000001&mcdnid=1002569&mid=426998802&nbs=1&nettype=0&oi=1946640525&orderid=0%2C3&os=mcdn&platform=pc&sign=e86b6a&traceid=trbsuOfwPijJKv_0_e_N&uipk=5&uparams=e%2Cuipk%2Cnbs%2Cdeadline%2Cgen%2Cos%2Coi%2Ctrid%2Cmid%2Cplatform&upsig=ba325e60ba697a0fd7c72f1bb12ff2a8"
	dmt := meta.NewDownloadMeta(url, "C:/Users/Ambitious/Downloads/1.mp4", url)
	dl := coredl.NewMp4SimpleDownloader()
	err := dl.Exec(dmt, func(current, total int64) {
		percent := float64(current) / float64(total) * 100
		log.Success(fmt.Sprintf("当前文件下载进度：%v/%v(%.2f%%)", current, total, percent))
	})
	if err != nil {
		t.Error(err)
	}
}
