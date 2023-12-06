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
	url := "https://pcvideoaliyun.titan.mgtv.com/c1/2019/04/12_0/B38C5A28076835FD2E24FE8879E5955C_20190412_1_1_3433_mp4/30AAB778F4D5F2D448905BD5761F834E.m3u8?arange=0&pm=guwl9aUQM49sIUf2AUk2ok_2ROO6oRf9MFr~pQrXKobjjXsjfkzrx0DnQ57jogMENwbCEQqoS~DjrdPJ_NHCueSKfdmVeZGmsaosamKmfWwzQTCku4OqYeyOoh0Ghs5C8xdpERjEdgleBeKA393fWpSC4aP2Vbi5TBqTLC49~vGGHs1K2g7AZl2ExGv2ZsB489euwll_xzZ~TQUpRCSCkS~Ik257W33lZyumTx8HqisEauoKyQ92YGFXN1HakxvdzOMIi1RwXIaPdhmS9xw14NnDxTvl8EgeVLHiEu6z~3xUQkyL7kNoFNRKDl3VUfIyg2uazn_3cIToyLTWGwc_IrJ9jnlyW559GCZWj8TbnHh7A02HJmN3IlI~WbKwjiyzyOMtmt2koM1e4kYCAh15A3GPes4~PmfcuYVLG2vy2gJQJt3D0e1yUBRpyJ5umFtkcmWN0~kIBYM-&mr=0SkZ6FKFEI9VOGbwBG~mtRg7YL5zflXtxqZR0LJuqPaCSOe5IIw8eVKc4zIndlQ6a_UQ_yTyijpWY~b3GJh9EFeKIOjfjK3qmJKzCW7L7Hmj8ecg0DqeJjF~Rd9mGqGpLR8LEUhq10NLXb1A7NWLvxawXC8Ym6R2zNe53shTyrVWjdEaHv4KovrSfhg~fC1C7KoiujAidWDv9WY4gWd17~dhTHlXR0i1ilhwtOsS2CajVmNovctoh2KhSCXLlHgfS~uW0a0BR_n9mXUpo16Swd9ZpMSmAyic5LlyjDY4vTd1zQ1ZY7LSbHg34JsiDgzPEafvZDZjdzHwztNzkZ~ObTo93f~AzzoDk0pTALA~m2HB1sXvA0emB5MmFs3vJ1afJ1PXu2xTfE5BcXoc3O~EHkJ8Teepu0qbAK1JwelTllQgcL4Brv6yCIx5qMTL6i_h62AXO0rQcg9aZgLmMkK05ySp1aDCnrYDjAETsYIkG55h6UFgmvLj8QSME992lgOkTBLDQWOvpeFY~avb9bpnhCxIp1mgc0qL64oZUn3~u_UzpRjgjNchlSz9dVXzYyoo0AHdPr9APgXQn44OH8t2ZkS20zoeM6ai~15cK35cXbWFa_JzEq7QthoVIhTxDSB3fPAmbFyXKkhPSa7bRmPslg--&uid=e4f3fabc8ec345b49c021c67e1c2a082&scid=25021&cpno=6i06rp&ruid=2fc9ce3d25c341be&sh=1"
	dmt := meta.NewDownloadMeta(url, "/Users/ambitious/Downloads/1.mp4", url)
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
