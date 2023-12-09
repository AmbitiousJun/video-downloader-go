// m3u8 视频下载
package coredl

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"video-downloader-go/internal/appctx"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/downloader/dlpool"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util"
	"video-downloader-go/internal/util/m3u8"
	"video-downloader-go/internal/util/myfile"
	"video-downloader-go/internal/util/myhttp"

	"github.com/pkg/errors"
)

const (
	TsFilenameFormat = "ts_%d.ts" // ts 文件格式
)

// m3u8 单协程下载器
type m3u8SimpleDownloader struct{}

// m3u8 多协程下载器
type m3u8MultiThreadDownloader struct{}

func (d *m3u8SimpleDownloader) Exec(dmt *meta.Download, handlerFunc ProgressHandler) error {
	return downloadM3U8(dmt, handlerFunc, false)
}

func (d *m3u8MultiThreadDownloader) Exec(dmt *meta.Download, handlerFunc ProgressHandler) error {
	return downloadM3U8(dmt, handlerFunc, true)
}

// 下载 m3u8 视频的核心逻辑
func downloadM3U8(dmt *meta.Download, handlerFunc ProgressHandler, multiThread bool) error {
	var current, total, currentBytes int64
	// 1 读取 ts 文件
	dmt.HeaderMap = myhttp.GenDefaultHeaderMapByUrl(dmt.HeaderMap, dmt.Link)
	tsMetas, err := m3u8.ReadTsUrls(dmt.Link, dmt.HeaderMap)
	if err != nil {
		return errors.Wrapf(err, "读取 ts 文件失败，file: %v", dmt.FileName)
	}
	total = int64(len(tsMetas))
	if total == 0 {
		return errors.New("读取到空 m3u8，下载任务终止")
	}
	handlerFunc(&Progress{
		Current:      current,
		Total:        total,
		CurrentBytes: currentBytes,
		TotalBytes:   currentBytes,
	})
	// 2 初始化临时文件夹
	tempDirPath, err := myfile.InitTempTsDir(dmt.FileName, config.G.Downloader.TsDirSuffix)
	if err != nil {
		return errors.Wrapf(err, "初始化临时 ts 文件夹失败，file: %v", dmt.FileName)
	}
	// 3 执行下载
	downloadTsMeta := func(tmt *m3u8.TsMeta) {
		var tmpErr error
		defer func() {
			err = util.AnyError(err, tmpErr)
		}()
		// 通过外部函数的 err 对象来传递错误
		tsPath := filepath.Join(tempDirPath, fmt.Sprintf(TsFilenameFormat, tmt.Index))
		var req *http.Request
		req, tmpErr = http.NewRequest(http.MethodGet, tmt.Url, nil)
		if tmpErr != nil {
			tmpErr = errors.Wrapf(tmpErr, "构造请求时出现异常：%v", dmt.FileName)
			return
		}
		for k, v := range dmt.HeaderMap {
			req.Header.Add(k, v)
		}
		var dn int64
		if dn, tmpErr = myhttp.DownloadWithRateLimit(req, tsPath); tmpErr != nil {
			tmpErr = errors.Wrapf(tmpErr, "分片下载异常：%v", dmt.FileName)
			return
		}
		// 每个分片下载完成的时候调用进度监听器
		atomic.AddInt64(&current, 1)
		atomic.AddInt64(&currentBytes, dn)
		handlerFunc(&Progress{
			Current:      current,
			Total:        total,
			CurrentBytes: currentBytes,
			TotalBytes:   currentBytes,
		})
	}
	if multiThread {
		err = util.AnyError(err, handleTsMetasMultiThread(tsMetas, downloadTsMeta))
	} else {
		handleTsMetasSimple(tsMetas, downloadTsMeta)
	}
	if err != nil {
		return errors.Wrap(err, "m3u8 下载失败")
	}
	// 4 合并文件
	if err = m3u8.Merge(tempDirPath); err != nil {
		return errors.Wrap(err, "合并 ts 文件失败")
	}
	return nil
}

// 单协程下载 ts 文件
func handleTsMetasSimple(tsMetas []*m3u8.TsMeta, downloadFunc func(*m3u8.TsMeta)) {
	if len(tsMetas) == 0 {
		return
	}
	for _, tmt := range tsMetas {
		downloadFunc(tmt)
	}
}

// 多协程下载 ts 文件
func handleTsMetasMultiThread(tsMetas []*m3u8.TsMeta, downloadFunc func(*m3u8.TsMeta)) (err error) {
	if len(tsMetas) == 0 {
		return nil
	}
	// 协程同步器用于同步多协程下载
	var wg sync.WaitGroup
	for _, tmt := range tsMetas {
		copyMt := tmt
		wg.Add(1)
		err = dlpool.SubmitDownload(func() {
			appctx.WaitGroup().Add(1)
			defer wg.Done()
			defer appctx.WaitGroup().Done()
			select {
			case <-appctx.Context().Done():
				return
			default:
				if err == nil {
					downloadFunc(copyMt)
				}
			}
		})
		if err != nil {
			return errors.Wrap(err, "协程池异常，请检查配置")
		}
	}
	// 等待所有协程运行完毕
	wg.Wait()
	return
}
