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
	list.OfferLast(meta.NewDownloadMeta("https://apd-vlive.apdcdn.tc.qq.com/defaultts.tc.qq.com/B_tRCdt2L6hl1ezG-aht1_p8Bh8lDqIF_3_hl_RJNvCqjSmaOVoJqwRvRqDldWh1xC/svp_50112/Kk0xxQEbWOiyIG_GbcrS_P2JsRJhklLCRlT9mQFF_rV_RYLVFmVnVNLNMGmoq_ubVbx3aDh8Vo_4FyEsBykMhUxkx5rm057TTfith8Oyu0GC9sL6rJt2tGxIF3ulqbD78IIOdKk4gsV2CR6k4DRg_MXMrg34rKmcnjQVnQFf-9_tFMxjKN2nwFpbKsiT2Y4zszIsSviY62ziwMO5mvB0xx7B96w8y-V4R9b0H9rn_I2b_rfJan-7aw/gzc_1000102_0b535qaaiaaaryaknhlyyzs4b3gdatqqaaca.f322016.ts.m3u8?ver=4", "测试", ""))

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
