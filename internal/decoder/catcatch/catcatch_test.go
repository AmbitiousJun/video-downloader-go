package catcatch_test

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/decoder/catcatch"
	"video-downloader-go/internal/util/mylog"
)

func genResults() []catcatch.CatCatchResult {
	src := `
	[
		{
			"action":"catCatchAddMedia",
			"url":"https://ltssjy.qq.com/B_tRCdt2L6hl1ezG-aht1_p-v1r89KLMpd5knwbvxIgh7iX-0x9w95_rx7NVLaVkg3/svp_50112/WDjuVsfBvkM_jmkTcjqPtU8Y_1tTCHbq9DHP4z1KXFiEspOqOkFgw7zw5K_TJLVL0e-SCCHRfvWR0kqWpIrV-HUW9VLz48v5vgbMfMuZ0pbiEJHm1_r1YydUmjJGlcRGma_xnS2t0vP2-7V4Pq69CL377bn63dzkGR--WPSseYi9g33D_cRy5dtbroAi9b9MPdsd2soB6trAM9Tvm4gnRzVUdGO4q9YRfBAtu6ZlAwVgx_zd-3Tzaw/gzc_1000102_0b53e4admaaad4afhtvwo5s4aj6dgyvaamsa.f322016.ts.m3u8?ver=4",
			"href":"https://v.qq.com/x/cover/mzc00200xkb43kw/l0048roeudh.html",
			"ext":"m3u8",
			"requestId":"17070302509791"
		},
		{
			"action":"catCatchAddMedia",
			"url":"https://207133493e769cbd1137a0616408676e830430c7f21199af.v.smtcdns.com/varietyts.tc.qq.com/ADC7Sl5SqmMhSxzFWrF23TnZ731hsJguxalhgxMosmTQ/B_tRCdt2L6hl1ezG-aht1_p2SpPqJPU8cOW-zO3KEMOeDiX-0x9w95_rx7NVLaVkg3/svp_50112/WDjuVsfBvkM_jmkTcjqPtU8Y_1tTCHbq9DHP4z1KXFiEspOqOkFgw7zw5K_TJLVL0e-SCCHRfvWR0kqWpIrV-HUW9VLz48v5vgbMfMuZ0pbiEJHm1_r1YydUmjJGlcRGma_xnS2t0vP2-7V4Pq69CL377bn63dzkGR--WPSseYi9g33D_cRy5dtbroAi9b9MPdsd2soB6trAM9Tvm4gnRzVUdGO4q9YRfBAtu6ZlAwVgx_zd-3Tzaw/gzc_1000102_0b53e4admaaad4afhtvwo5s4aj6dgyvaamsa.f322016.ts.m3u8?ver=4",
			"href":"https://v.qq.com/x/cover/mzc00200xkb43kw/l0048roeudh.html",
			"ext":"m3u8",
			"requestId":"17070302509802"
		},
		{
			"action":"catCatchAddMedia",
			"url":"https://f8b9374e22d79d32c011a334f7df8065fdaa3694e4312f64.v.smtcdns.com/varietyts.tc.qq.com/ADC7Sl5SqmMhSxzFWrF23TnZ731hsJguxalhgxMosmTQ/B_tRCdt2L6hl1ezG-aht1_p2SpPqJPU8cOW-zO3KEMOeDiX-0x9w95_rx7NVLaVkg3/svp_50112/WDjuVsfBvkM_jmkTcjqPtU8Y_1tTCHbq9DHP4z1KXFiEspOqOkFgw7zw5K_TJLVL0e-SCCHRfvWR0kqWpIrV-HUW9VLz48v5vgbMfMuZ0pbiEJHm1_r1YydUmjJGlcRGma_xnS2t0vP2-7V4Pq69CL377bn63dzkGR--WPSseYi9g33D_cRy5dtbroAi9b9MPdsd2soB6trAM9Tvm4gnRzVUdGO4q9YRfBAtu6ZlAwVgx_zd-3Tzaw/gzc_1000102_0b53e4admaaad4afhtvwo5s4aj6dgyvaamsa.f322016.ts.m3u8?ver=4",
			"href":"https://v.qq.com/x/cover/mzc00200xkb43kw/l0048roeudh.html",
			"ext":"m3u8",
			"requestId":"17070302509803"
		},
		{
			"action":"catCatchAddMedia",
			"url":"https://apd-vlive.apdcdn.tc.qq.com/defaultts.tc.qq.com/B_tRCdt2L6hl1ezG-aht1_p3roOfHvqyMUBAjNELY83IziX-0x9w95_rx7NVLaVkg3/svp_50112/WDjuVsfBvkM_jmkTcjqPtU8Y_1tTCHbq9DHP4z1KXFiEspOqOkFgw7zw5K_TJLVL0e-SCCHRfvWR0kqWpIrV-HUW9VLz48v5vgbMfMuZ0pbiEJHm1_r1YydUmjJGlcRGma_xnS2t0vP2-7V4Pq69CL377bn63dzkGR--WPSseYi9g33D_cRy5dtbroAi9b9MPdsd2soB6trAM9Tvm4gnRzVUdGO4q9YRfBAtu6ZlAwVgx_zd-3Tzaw/gzc_1000102_0b53e4admaaad4afhtvwo5s4aj6dgyvaamsa.f322016.ts.m3u8?ver=4",
			"href":"https://v.qq.com/x/cover/mzc00200xkb43kw/l0048roeudh.html",
			"ext":"m3u8",
			"requestId":"17070302509804"
		}
	]
	`
	results := []catcatch.CatCatchResult{}
	json.Unmarshal([]byte(src), &results)
	return results
}

func TestPrintResult(t *testing.T) {
	results := genResults()
	catcatch.PrintResult(results, func(line string) {
		log.Println(line)
	})
}

func TestReadJsonConfig(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()

	if err := config.Load("../../../config/config.yml"); err != nil {
		t.Error(err)
		return
	}

	td := new(catcatch.TxDecoder)

	cookies := td.ReadCookiesFromConfig()
	for _, cookie := range cookies {
		mylog.Successf("%v", cookie)
	}
}

func TestResultSelector(t *testing.T) {
	defer appctx.WaitGroup().Wait()
	defer appctx.CancelFunc()()

	// 模拟控制台输入
	originStdIn := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() {
		os.Stdin = originStdIn
		w.Close()
	}()
	w.Write([]byte("\n"))
	w.Write([]byte("17070302509804\n"))

	rs := catcatch.NewResultSelector(genResults())
	res, err := rs.Select()
	if err != nil {
		t.Error(err)
		return
	}
	mylog.Successf("用户选择的结果是: %v", res)
}
