// 下载器配置
package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"video-downloader-go/internal/util/mylog"
	"video-downloader-go/internal/util/mytokenbucket"

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
	DownloadSimple      = "simple"       // 单线程下载
	DownloadMultiThread = "multi-thread" // 多线程下载
)

const (
	RateLimitMaxValueKBPS         = float64(math.MaxInt32) / 2 / 1024 // kbps 最大下载速率
	RateLimitMinValueKBPS float64 = 1.0 * 10                          // kbps 最小下载速率
	RateLimitMaxValueMBPS         = RateLimitMaxValueKBPS / 1024      // mbps 最大下载速率
	RateLimitMinValueMBPS float64 = 0.1                               // mbps 最小下载速率
)

// 检查下载器配置是否正确
func checkDownloaderConfig() error {
	cfg := &G.Downloader
	cfg.DownloadDir = strings.TrimSpace(cfg.DownloadDir)
	if cfg.DownloadDir == "" {
		return errors.New("下载目录不能为空")
	}
	if strings.TrimSpace(cfg.Use) == "" || !downloadTypeValid(cfg.Use) {
		mylog.Warn("没有配置下载类型或配置错误，默认使用多线程下载")
		cfg.Use = DownloadMultiThread
	}
	if cfg.TaskThreadCount <= 0 {
		mylog.Warn("没有配置下载任务处理线程数或配置错误，使用默认值：2")
		cfg.TaskThreadCount = 2
	}
	if cfg.DlThreadCount <= 0 {
		mylog.Warn("没有配置下载线程数或配置错误，使用默认值：32")
		cfg.DlThreadCount = 32
	}
	if strings.TrimSpace(cfg.TsDirSuffix) == "" {
		mylog.Warn("没有配置临时 ts 目录后缀或配置错误，使用默认值：temp_ts_files")
		cfg.TsDirSuffix = "temp_ts_files"
	}
	// 默认速率是 5mbps
	var err error
	var rate float64 = 5 * 1024 * 1024
	kbps, mbps := "kbps", "mbps"
	if cfg.RateLimit = strings.TrimSpace(cfg.RateLimit); cfg.RateLimit != "" {
		if cfg.RateLimit == "-1" {
			rate = RateLimitMaxValueMBPS * 1024 * 1024
		} else if strings.HasSuffix(cfg.RateLimit, kbps) {
			val := cfg.RateLimit[:len(cfg.RateLimit)-len(kbps)]
			rate, err = checkKbpsRateLimit(val)
			if err != nil {
				return errors.Wrap(err, "检查下载器配置时出现异常")
			}
			mylog.Successf("下载速率限制：%.1f%v", rate, kbps)
			rate *= 1024
		} else if strings.HasSuffix(cfg.RateLimit, mbps) {
			val := cfg.RateLimit[:len(cfg.RateLimit)-len(mbps)]
			rate, err = checkMbpsRateLimit(val)
			if err != nil {
				return errors.Wrap(err, "检查下载器配置时出现异常")
			}
			mylog.Successf("下载速率限制：%.1f%v", rate, mbps)
			rate *= 1024 * 1024
		} else {
			mylog.Warn("没有配置限速或者配置出错，启用默认的速率限制：5mbps")
		}
	}
	// 初始化令牌桶
	tokenBucket, err := mytokenbucket.NewTokenBucket(int64(rate))
	if err != nil {
		return errors.Wrap(err, "初始化速率限制令牌桶时出现异常")
	}
	mytokenbucket.GlobalBucket = tokenBucket
	return nil
}

// 检查 kbps 速率是否合法
// @param val 用户输入的速率
// @return 如果合法，转换成 float64 类型并返回，不合法则返回 error
func checkKbpsRateLimit(val string) (float64, error) {
	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, errors.Wrapf(err, "转换下载速率异常：%v", val)
	}
	if res < RateLimitMinValueKBPS || res > RateLimitMaxValueKBPS {
		return -1, errors.New(fmt.Sprintf("速率限制范围（kbps）：[%.1f, %.1f]", RateLimitMinValueKBPS, RateLimitMaxValueKBPS))
	}
	return res, nil
}

// 检查 mbps 速率是否合法
// @param val 用户输入的速率
// @return 如果合法，转换成 float64 类型并返回，不合法则返回 error
func checkMbpsRateLimit(val string) (float64, error) {
	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, errors.Wrapf(err, "转换下载速率异常：%v", val)
	}
	if res < RateLimitMinValueMBPS || res > RateLimitMaxValueMBPS {
		return -1, errors.New(fmt.Sprintf("速率限制范围（mbps）：[%.1f, %.1f]", RateLimitMinValueMBPS, RateLimitMaxValueMBPS))
	}
	return res, nil
}

// 检查下载类型是否合法
// @param use 下载器类型
// @return 是否合法
func downloadTypeValid(use string) bool {
	return use == DownloadSimple || use == DownloadMultiThread
}
