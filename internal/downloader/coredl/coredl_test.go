package coredl_test

import (
	"fmt"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
)

// 测试下载 m3u8
func TestDownloadM3U8(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	config.Load("../../../config/config.yml")
	dlpool.InitFromGlobalConfig()
	defer dlpool.ReleaseAll()
	url := "https://apd-vlive.apdcdn.tc.qq.com/defaultts.tc.qq.com/B_tRCdt2L6hl1ezG-aht1_p_DeuKOAWdX28Xl7uGkLQ_xjyNAfbZapX9DVCYN9S0I7/svp_50112/fcrR6H4oEIWkJ5-Iulz5nXsOk54a7ZVDd80sDIOngcgPWvK-80dTvsyAqLe53QQ01x9RH7yJZ4uLYz79D_qCBxHJCdH3p2Yhruul7mtJ7fe_vI7SnTpsbdYFs3DcnqqJe1u6u73G78UCowCtNYOQOBV-aFValIOdz7BtjJbFaY2bKJmoikMLmizsnnSYtus4KSPwT0n_BCbhhRvrVwiWLcvH2m_vzRPgfyM5PJQ6nVNe_B3htBFV4Q/gzc_1000102_0b53aeacaaaa2maj7a3pcbs4aaodeagaajca.f322016.ts.m3u8?ver=4"
	dmt := meta.NewDownloadMeta(url, "C:/Users/Ambitious/Downloads/1.mp4", url)
	dl := coredl.NewM3U8MultiThread()
	err := dl.Exec(dmt, func(p *coredl.Progress) {
		percent := float64(p.Current) / float64(p.Total) * 100
		mylog.Success(fmt.Sprintf("当前文件下载进度：%v/%v(%.2f%%)，已下载：%vbytes，总共：%vbytes", p.Current, p.Total, percent, p.CurrentBytes, p.TotalBytes))
	})
	if err != nil {
		t.Error(err)
	}
}

// 测试下载 MP4
func TestDownloadMp4(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	config.Load("../../../config/config.yml")
	dlpool.InitFromGlobalConfig()
	defer dlpool.ReleaseAll()
	url := "https://xy218x87x208x202xy.mcdn.bilivideo.cn:8082/v1/resource/1313883886-1-100047.m4s?agrr=0&build=0&buvid=61104810-23A2-5AFC-7FC9-99DA27C7865795586infoc&bvc=vod&bw=38352&deadline=1701749605&e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M%3D&f=u_0_0&gen=playurlv2&logo=A0000400&mcdnid=11000365&mid=0&nbs=1&nettype=0&oi=1946640849&orderid=0%2C3&os=mcdn&platform=pc&sign=504a61&traceid=trMtoYhTxlnnxZ_0_e_N&uipk=5&uparams=e%2Cuipk%2Cnbs%2Cdeadline%2Cgen%2Cos%2Coi%2Ctrid%2Cmid%2Cplatform&upsig=8b02362065afe98cb27567d30cc69001"
	dmt := meta.NewDownloadMeta(url, "C:/Users/Ambitious/Downloads/1.mp4", url)
	dl := coredl.NewMp4MultiThread()
	err := dl.Exec(dmt, func(p *coredl.Progress) {
		percent := float64(p.Current) / float64(p.Total) * 100
		mylog.Success(fmt.Sprintf("当前文件下载进度：%v/%v(%.2f%%)，已下载：%vbytes，总共：%vbytes", p.Current, p.Total, percent, p.CurrentBytes, p.TotalBytes))
	})
	if err != nil {
		t.Error(err)
	}
}
