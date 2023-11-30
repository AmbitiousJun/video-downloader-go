package coredl

import (
	"math"
	"math/rand"
	"net/http"
	"time"
	"video-downloader-go/internal/meta"
	"video-downloader-go/internal/util/myhttp"
	"video-downloader-go/internal/util/mymath"

	"github.com/pkg/errors"
)

const (
	SplitCount = 64 // 一个 MP4 文件至少分割成多少份
)

// mp4 单协程下载
type mp4SimpleDownloader struct{}

// mp4 多协程下载
type mp4MultiThreadDownloader struct{}

// 一个 mp4 任务分割出来的分片任务
type unitTask struct {
	from int64 // 起始字节（闭）
	to   int64 // 终止字节（开）
}

// 使用 mp4 单协程下载时，current 和 total 分别是已下载字节数和总字节数
func (d *mp4SimpleDownloader) Exec(meta *meta.Download, progressListener func(current, total int64)) error {
	var current, total int64 = 0, 0
	// 1 获取文件总大小
	ranges, err := myhttp.GetRequestRangesFrom(meta.Link, http.MethodGet, myhttp.GenDefaultHeaderMapByUrl(nil, meta.Link), 0)
	if err != nil || ranges[1] <= 0 {
		return errors.Wrap(err, "无法获取文件总大小或文件为空")
	}
	total = ranges[1]
	// 2 分片
	tasks := initUnitTasks(total)
	// 3 循环分片进行下载
	return nil
}

func (d *mp4MultiThreadDownloader) Exec(meta *meta.Download, progressListener func(current, total int64)) error {
	return nil
}

// 初始化下载任务分片列表
func initUnitTasks(fileTotalSize int64) []*unitTask {
	size := int64(math.Ceil(float64(fileTotalSize) / SplitCount))
	// 每个分片大小 2~4 MB
	var baseSize int64 = 2 * 1024 * 1024
	var i int64
	res := []*unitTask{}
	for ; i < SplitCount; i++ {
		curSize := int64(mymath.Min(size, fileTotalSize-i*size))
		from := i * curSize
		rand.Seed(time.Now().UnixNano())
		for curSize > 2*baseSize {
			random := int64(float64(baseSize) * rand.Float64())
			to := from + baseSize + random
			res = append(res, &unitTask{from: from, to: to})
			curSize -= (baseSize + random)
			from = to
		}
		if curSize > 0 {
			res = append(res, &unitTask{from: from, to: from + curSize})
		}
	}
	return res
}
