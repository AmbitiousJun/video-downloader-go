package downloader_test

import (
	"fmt"
	"sync"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
)

// 测试下载监听器能否正常运行
func TestUseListenerToDownload(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()

	config.Load("../../config/config.yml")

	list := meta.TaskDeque[meta.Download]{}
	list.OfferLast(meta.NewDownloadMeta("https://pcvideotx.titan.mgtv.com/c1/2023/12/08_0/0FD4F89CBCABBBCAF405DEA01A3DCAE2_20231208_1_1_1262_mp4/BDFE57CBA270DB03F7FE6E37C011380B.m3u8?arange=0&pm=f448M50P423B~oMhswb8YGZKBVgz8ed4OZDY9KHX~TIhuypN99R9xBy4mF~mZJQptrBbwVqPCafoNtDqVboGaXIBZmE_s2eDLxSFv2HlBPa5JABcVBKsClUk063nJ4BXx8dLMQGkGwx5WFdB_3G6s4ZuYlbCNdUaGYe1xsr5T6Nt0VuugaO3MosmRpxZgLa1u35l8m2KLShhTboYzjVyZZKezECk5uB2lVBqAG40tKowAe090H4NTGsAuHR46DlbmKYCCtnlk91~7KzdryCoKco1wrcUGf_pvI1wQXk7DDdafqpKGasx8xLHl5TnbWFlTnUStPWfui~OMv2gmb3M~43lmPINaItnzmdaK6NHPQoPYQL_4oT9SD3vOw~eIH8s~MYzCanTPEwmGN2HC_LpbD6KfXh2XjL1gmkeZZJN~mYQ1j213CoF8n4vzWyfp5FvQSLwpg--&mr=cvfrkoKGkW_QN5KlICuFhh8459MUzAdFnSbFpItJU1azjjXmbUnP1I_wLaT045xU_jFzoXNzfQ1kpRPEGvomIwsoVlRBjWPbJYxrfNb27FYOf2wrg93e06VCxx3Qh6qZrboOLxN3Wg8~gZo63dAw14Vuejwr5xfbTHhWKkQVYhxiP084fnHaRkI4UwEpjSUAcqfEF9tV2QOWCkVH41_YPeaJttQLyVTbN5rsnBCnHVef6v6RjdY3T2N0VmolJLsQVguE8PdRR7m6bzWqtjT6Qv_YJPAm68A0iQ7fsKuqqw044RyWJSwtlbOl0nUNj9KU7UGGI~nhSEvAsUhQNf3MR~yawTnUj34YnEUYZrWWrTEg3CCQO29tg_Yw7j1SKPI5Fv3DtZnSaCai0umahdaP2rVzHK75tKavQhIETyEjUyknc7yVZfUv0ZckPAW_8vkiKjgEfeD4EfJtU9B6LN7DZpCzxsdWN4FW8tb3kwRk5IDefTvYyEH9YXd3toO_ja4zGEGIjwGZRZWbToV2yeCNJamaAsEJDAYAsop3cMJGdZQy9TkcZFR6AC2gdqkI_iWnk_n~rmDDiHivwGXYlnvIGc~iQz~~5vzylXC0ajV8jAOscWD0pvcRxdN9K1A4X7nqPt0z0fW5ol0cFsyr&uid=e4f3fabc8ec345b49c021c67e1c2a082&scid=25015&cpno=6i06rp&ruid=691147079a7c4311&sh=1", "测试", ""))

	var wg sync.WaitGroup
	wg.Add(1)
	downloader.ListenAndDownload(&list, func() {
		mylog.Success("成功下载完成一个任务")
		wg.Done()
	}, func(dmt *meta.Download) {
		mylog.Error(fmt.Sprintf("下载失败了, %v", dmt))
		wg.Done()
	})
	wg.Wait()

}
