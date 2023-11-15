// 下载器配置
package config

import (
	"math"
	"strings"
	"video-downloader-go/src/util/log"

	"github.com/pkg/errors"
)

type Downloader struct {
	Use             string `yaml:"use"`               // 要使用哪个下载器，可选值：simple, multi-thread
	TaskThreadCount int    `yaml:"task-thread-count"` // 处理下载任务的线程个数
	DlThreadCount   int    `yaml:"dl-thread-count"`   // 多线程下载的线程个数
	DownloadDir     string `yaml:"download-dir"`      // 视频文件下载位置
	TsDirSuffix     string `yaml:"ts-dir-suffix"`     // 暂存 ts 文件的目录后缀
	RateLimit       string `yaml:"rate-limit"`        // 下载限速，两种单位可选：mbps，kbps，-1 则不限速
}

const (
	SimpleDownload      = "simple"       // 单线程下载
	MultiThreadDownload = "multi-thread" // 多线程下载
)

const (
	RateLimitMaxValueKBPS         = float64(math.MaxInt32) / 2 / 1024 // kbps 最大下载速率
	RateLimitMinValueKBPS float64 = 1.0 * 10                          // kbps 最小下载速率
	RateLimitMaxValueMBPS         = RateLimitMaxValueKBPS / 1024      // mbps 最大下载速率
	RateLimitMinValueMBPS float64 = 0.1                               // mbps 最小下载速率
)

// 检查下载器配置是否正确
func checkDownloaderConfig() error {
	cfg := GlobalConfig.Downloader
	cfg.DownloadDir = strings.TrimSpace(cfg.DownloadDir)
	if cfg.DownloadDir == "" {
		return errors.New("下载目录不能为空")
	}
	if strings.TrimSpace(cfg.Use) == "" || !downloadTypeValid(cfg.Use) {
		log.Warn("没有配置下载类型或配置错误，默认使用多线程下载")
		cfg.Use = MultiThreadDownload
	}
	if cfg.TaskThreadCount <= 0 {
		log.Warn("没有配置下载任务处理线程数或配置错误，使用默认值：2")
		cfg.TaskThreadCount = 2
	}
	if cfg.DlThreadCount <= 0 {
		log.Warn("没有配置下载线程数或配置错误，使用默认值：32")
		cfg.DlThreadCount = 32
	}
	if strings.TrimSpace(cfg.TsDirSuffix) == "" {
		log.Warn("没有配置临时 ts 目录后缀或配置错误，使用默认值：temp_ts_files")
		cfg.TsDirSuffix = "temp_ts_files"
	}
	// TODO: 初始化令牌桶
	return nil
}

// 检查下载类型是否合法
// @param use 下载器类型
// @return 是否合法
func downloadTypeValid(use string) bool {
	return use == SimpleDownload || use == MultiThreadDownload
}
