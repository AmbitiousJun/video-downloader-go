package decoder

import (
	"time"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/mylog"
)

// 解析器通用接口
type D interface {
	// 解析并获取下载链接列表
	// 对于 youtube-dl 解析器，有可能会返回两条链接，因为部分站点的解析结果是音视频分开的
	// 其他解析器通常只返回一条链接
	FetchDownloadLinks(url string) ([]string, error)
}

// 解析成功的处理函数
type DecodeSuccessHandler func(*meta.Download)

func ListenAndDecode(list *meta.TaskDeque[meta.Video], decodeSuccess DecodeSuccessHandler) {
	mylog.Info("开始监听解析列表")
	go func() {
		ticker := NewGrowableTicker(10, 5*60, 0.2)
		for {
			// 没有任务处理时，每阻塞 2 秒检查一次任务列表
			for list.Empty() {
				time.Sleep(time.Second * 2)
			}

			// 获取链接地址
			vmt := list.PollFirst()
			mylog.Infof("识别到解析任务，标题：%s，源地址：%s", vmt.Name, vmt.Url)

			// 判断解析类型
			use := config.G.Decoder.CustomUse(vmt.Url)
			dcd := GetDecoder(use)

			if use == config.DecoderNone {
				// 不需要解析，url 直接就是下载链接
				decodeSuccess(meta.NewDownloadMeta(vmt.Url, vmt.Name, vmt.Url))
			} else if use == config.DecoderYoutubeDl {
				dmt, err := useYoutubeDlDecode(dcd, vmt)
				if err != nil {
					mylog.Errorf("视频下载地址解析失败：%v，重新加入任务列表", err)
					list.OfferLast(vmt)
				}
				decodeSuccess(dmt)
			}

			// 通常情况下，解析任务处理速率远高于下载任务
			// 所以这里阻塞一段较长的时间，避免解析过快
			time.Sleep(time.Second * ticker.Next())
		}
	}()
}

// useYoutubeDlDecode 调用 youtube-dl 解析器来解析下载地址
func useYoutubeDlDecode(dcd D, vmt *meta.Video) (*meta.Download, error) {
	mylog.Infof("开始解析视频，文件名：%s, 源地址：%s", vmt.Name, vmt.Url)
	links, err := dcd.FetchDownloadLinks(vmt.Url)
	if err != nil {
		return nil, err
	}
	mylog.Successf("解析成功，已添加到下载列表，文件名：%s，下载地址：%s", vmt.Name, links)
	return meta.NewYtDlDownloadMeta(links, vmt.Name, vmt.Url), nil
}
