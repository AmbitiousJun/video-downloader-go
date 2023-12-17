package coredl_test

import (
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/coredl"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
)

// 测试下载 m3u8
func TestDownloadM3U8(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()
	config.Load("../../../config/config.yml")
	url := "https://pcvideotx.titan.mgtv.com/c1/2023/11/22_0/AF45A8BC4119CC876C90BE52B447A3FC_20231122_1_1_2806_mp4/0EEDB4725C221FB785479F2272C2EE20.m3u8?arange=0&pm=YASFK_msQgcyAFo0KKjsyWn5hq6CSi7rWFyBdg_PvcCuV1MMnOjOvTJY6BWL26ryWrblg6kYPZ_OWgNZgikyibfq1oipK4Vr679umJte4v1TI_pio5zcJGbwg~Iooxmtu3FPd4NDKxfqKRTz7npygKv6uK5jewJr9H3bujEisXeex3puBtMdYGAfWEvNYmzlsyfjZ~ZUO2nCRGfRtJuOtfXQjJ3bvjgtGT99KspsEnIP5taCi5T_b0fskAc5xVttWKLzkoXzgnFUl4sYzPtgUgIoYmIQKQCS_3XE6Puo6Juy~4BUFEQ8FpOCgrbySyyOnOch8NjQ939IG5RwMkzRTmMalHJtbxrhdF49ohke4UvDanFypQtzC9ye760ZFJXSm3IW928k8_b7ZY6tyL0rKI24JxcHMgMbuAThDKveY_4RVjMBpIH9AwqT_F6PJXR~frucNA--&mr=cvfrkoKGkW_QN5KlICuFhh8459MUzAdFnSbFpItJU1azjjXmbUnP1I_wLaT045xU_jFzoXNzfQ1kpRPEGvomIwsoVlRBjWPbJYxrfNb27FYOf2wrg93e06VCxx3Qh6qZrboOLxN3Wg8~gZo63dAw14Vuejwr5xfbTHhWKkQVYhxiP084fnHaRkI4UwEpjSUAcqfEF9tV2QOWCkVH41_YPeaJttQLyVTbN5rsnBCnHVef6v6RjdY3T2N0VmolJLsQVguE8PdRR7m6bzWqtjT6Qv_YJPAm68A0iQ7fsKuqqw044RyWJSwtlbOl0nUNj9KU7UGGI~nhSEvAsUhQNf3MR~yawTnUj34YnEUYZrWWrTEg3CCQO29tg_Yw7j1SKPI5Fv3DtZnSaCai0umahdaP2rVzHK75tKavQhIETyEjUyknc7yVZfUv0ZckPAW_8vkiKjgEfeD4EfJtU9B6LN7DZpCzxsdWN4FW8tb3kwRk5IDefTvYyEH9YXd3toO_ja4zGEGIjwGZRZWbToV2yeCNJamaAsEJDAYAsop3cMJGdZQy9TkcZFR6AC2gdqkI_iWnk_n~rmDDiHivwGXYlnvIGc~iQz~~5vzylXC0ajV8jAOscWD0pvcRxdN9K1A4X7nqPt0z0fW5ol0cFsyr&uid=e4f3fabc8ec345b49c021c67e1c2a082&scid=25015&cpno=6i06rp&ruid=c7499212b13a4859&sh=1"
	dmt := meta.NewDownloadMeta(url, "/Users/ambitious/Downloads/1.mp4", url)
	dl := coredl.NewM3U8MultiThread()
	err := dl.Exec(dmt, func(p *coredl.Progress) {
		percent := float64(p.Current) / float64(p.Total) * 100
		mylog.Successf("当前文件下载进度：%v/%v(%.2f%%)，已下载：%vbytes，总共：%vbytes", p.Current, p.Total, percent, p.CurrentBytes, p.TotalBytes)
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
	url := "https://xy182x54x114x202xy.mcdn.bilivideo.cn:8082/v1/resource/1356354534-1-100113.m4s?agrr=0&build=0&buvid=60418CB0-A8FC-C4A0-8AC6-1BCC9F4A707090899infoc&bvc=vod&bw=44328&deadline=1702053896&e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M%3D&f=u_0_0&gen=playurlv2&logo=A0000400&mcdnid=11000365&mid=12151031&nbs=1&nettype=0&oi=2032357081&orderid=0%2C3&os=mcdn&platform=pc&sign=ca38a1&traceid=trFjKktnkYuuLH_0_e_N&uipk=5&uparams=e%2Cuipk%2Cnbs%2Cdeadline%2Cgen%2Cos%2Coi%2Ctrid%2Cmid%2Cplatform&upsig=60e4dff25b7090ea14234eedb4c0712e"
	dmt := meta.NewDownloadMeta(url, "/Users/ambitious/Downloads/1.mp4", url)
	dl := coredl.NewMp4MultiThread()
	err := dl.Exec(dmt, func(p *coredl.Progress) {
		percent := float64(p.Current) / float64(p.Total) * 100
		mylog.Successf("当前文件下载进度：%v/%v(%.2f%%)，已下载：%vbytes，总共：%vbytes", p.Current, p.Total, percent, p.CurrentBytes, p.TotalBytes)
	})
	if err != nil {
		t.Error(err)
	}
}
