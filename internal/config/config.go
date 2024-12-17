package config

import (
	"log"
	"os"
	"strings"
	"video-downloader-go/internal/lib/ffmpeg"
	"video-downloader-go/internal/lib/ytdlp"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var FfmpegPath string
var YoutubeDlPath string

type Config struct {
	Downloader Downloader     `yaml:"downloader"` // 下载器
	Transfer   Transfer       `yaml:"transfer"`   // 转换器
	Decoder    Decoder        `yaml:"decoder"`    // 解析器
	Customs    []CustomConfig `yaml:"customs"`    // 定制化配置
}

// 全局配置对象
var G = &Config{}

// 全局加载配置
func Load(configFilePath string) error {
	if configFilePath = strings.TrimSpace(configFilePath); len(configFilePath) == 0 {
		configFilePath = "config/config.yml"
	}

	// 1 读取整个配置文件
	fileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return errors.Wrap(err, "读取配置文件失败")
	}

	// 2 读取配置到 Config 结构中
	if err = yaml.Unmarshal(fileBytes, G); err != nil {
		return errors.Wrap(err, "读取配置文件失败")
	}

	// 3 读取依赖路径
	readDependencyPaths()

	// 4 检查下载器配置
	if err = checkDownloaderConfig(); err != nil {
		return errors.Wrap(err, "下载器配置异常")
	}

	// 5 检查转换器配置
	if err = checkTransferConfig(); err != nil {
		return errors.Wrap(err, "转换器配置异常")
	}

	// 6 检查解析器配置
	if err = checkDecoderConfig(); err != nil {
		return errors.Wrap(err, "解析器配置异常")
	}

	// 7 检查定制化配置
	if err = checkCustomConfig(); err != nil {
		return errors.Wrap(err, "定制化配置异常")
	}
	return nil
}

// 读取依赖路径地址
func readDependencyPaths() {
	if err := ytdlp.AutoDownloadExec(); err != nil {
		log.Panicf("yt-dlp 自动下载失败: %v, 请尝试重新运行程序或手动下载", err)
	}
	YoutubeDlPath = ytdlp.ExecPath()

	if err := ffmpeg.AutoDownloadExec(); err != nil {
		log.Panicf("ffmpeg 自动下载失败: %v, 请尝试重新运行程序或手动下载", err)
	}
	FfmpegPath = ffmpeg.ExecPath()
}
