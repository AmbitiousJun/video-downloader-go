package decoder

import (
	"errors"
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
)

// 解析器通用接口
type D interface {
	// 解析并获取下载链接列表
	// 对于 youtube-dl 解析器, 有可能会返回两条链接, 因为部分站点的解析结果是音视频分开的
	// 其他解析器通常只返回一条链接
	FetchDownloadLinks(url string) ([]string, error)
}

// 解析成功的处理函数
type DecodeSuccessHandler func(*meta.Download)

// 解析完成后将下载数据构建成 DownloadMeta
type DownloadMetaBuilder func([]string, *meta.Video) *meta.Download

func ListenAndDecode(list *meta.TaskDeque[meta.Video], decodeSuccess DecodeSuccessHandler) {
	mylog.Info("开始监听解析列表")
	go func() {
		ticker := NewGrowableTicker(30, 5*60, 0.17)
	out:
		for {
			// 没有任务处理时, 每阻塞 2 秒检查一次任务列表
			for list.Empty() {
				time.Sleep(time.Second * 2)
			}

			// 获取链接地址
			vmt := list.PollFirst()
			mylog.Infof("识别到解析任务, 标题：%s, 源地址：%s", vmt.Name, vmt.Url)
			vmt.LogBar.DecodeHint("解析中...")

			// 判断解析类型
			use := config.G.Decoder.CustomUse(vmt.Url)
			maxRetry := config.G.Decoder.CustomMaxRetry(vmt.Url)
			currentTry := 1
			dcd := GetDecoder(use)
			var dmt *meta.Download
			var decodeErr error

			for currentTry <= maxRetry {
				switch use {
				case config.DecoderNone:
					dmt, decodeErr = meta.NewDownloadMeta(vmt.Url, vmt.Name, vmt.Url), nil
				case config.DecoderYoutubeDl:
					dmt, decodeErr = useYoutubeDlDecode(dcd, vmt)
				case config.DecoderCatCatchTx:
					dmt, decodeErr = useCatCatchTxDecode(dcd, vmt)
				default:
					decodeErr = errors.New("不支持的解析器类型")
				}

				if decodeErr == nil {
					vmt.LogBar.WaitingHint("解析完成, 等待下载")
					dmt.LogBar = vmt.LogBar
					decodeSuccess(dmt)

					// 通常情况下, 解析任务处理速率远高于下载任务
					// 所以这里阻塞一段较长的时间, 避免解析过快
					if len(downloader.CanDownloadChan) != 0 {
						<-downloader.CanDownloadChan
					}
					select {
					case <-sleep2Channel(time.Second * ticker.Next()):
					case <-downloader.CanDownloadChan:
					}

					continue out
				}
				currentTry++
			}
			vmt.LogBar.ErrorHint("解析失败")
			mylog.Errorf("视频下载地址解析失败: %v", decodeErr)
		}
	}()
}

// useCatCatchTxDecode 调用 cat-catch:tx 解析器来解析下载地址
func useCatCatchTxDecode(dcd D, vmt *meta.Video) (*meta.Download, error) {
	return decode2Dmt(dcd, vmt, func(s []string, v *meta.Video) *meta.Download {
		return meta.NewDownloadMeta(s[0], v.Name, v.Url)
	})
}

// useYoutubeDlDecode 调用 youtube-dl 解析器来解析下载地址
func useYoutubeDlDecode(dcd D, vmt *meta.Video) (*meta.Download, error) {
	return decode2Dmt(dcd, vmt, func(s []string, v *meta.Video) *meta.Download {
		return meta.NewYtDlDownloadMeta(s, v.Name, v.Url)
	})
}

// decode2Dmt 通用的解析逻辑, 解析完成后, 使用 dmtBuilder 构建成 DownloadMeta 返回
func decode2Dmt(dcd D, vmt *meta.Video, dmtBuilder DownloadMetaBuilder) (*meta.Download, error) {
	mylog.Infof("开始解析视频, 文件名：%s, 源地址：%s", vmt.Name, vmt.Url)
	links, err := dcd.FetchDownloadLinks(vmt.Url)
	if err != nil {
		return nil, err
	}
	mylog.Successf("解析成功, 已添加到下载列表, 文件名：%s, 下载地址：%s", vmt.Name, links)
	return dmtBuilder(links, vmt), nil
}

// sleep2Channel 将线程睡眠转换成通道读取形式
func sleep2Channel(duration time.Duration) <-chan struct{} {
	ch := make(chan struct{}, 1)

	go func() {
		time.Sleep(duration)
		ch <- struct{}{}
		close(ch)
	}()

	return ch
}
