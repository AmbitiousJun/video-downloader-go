package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var FfmpegPath string
var YoutubeDlPath string

type Config struct {
	Downloader Downloader `yaml:"downloader"` // 下载器
	Transfer   Transfer   `yaml:"transfer"`   // 转换器
	Decoder    Decoder    `yaml:"decoder"`    // 解析器
	Os         string     `yaml:"os"`         // 运行操作系统
}

// 全局配置对象
var GlobalConfig = &Config{}

// 全局加载配置
func Load() error {
	// 1 读取整个配置文件
	fileBytes, err := ioutil.ReadFile("config/config.yml")
	if err != nil {
		return errors.Wrap(err, "读取配置文件失败")
	}
	// 2 读取配置到 Config 结构中
	err = yaml.Unmarshal(fileBytes, GlobalConfig)
	if err != nil {
		return errors.Wrap(err, "读取配置文件失败")
	}
	// 3 读取依赖路径
	readDependencyPaths()
	// 4 检查下载器配置
	err = checkDownloaderConfig()
	if err != nil {
		return errors.Wrap(err, "下载器配置异常")
	}
	// 5 检查转换器配置
	err = checkTransferConfig()
	if err != nil {
		return errors.Wrap(err, "转换器配置异常")
	}
	// 6 检查解析器配置
	err = checkDecoderConfig()
	if err != nil {
		return errors.Wrap(err, "解析器配置异常")
	}
	return nil
}

// 读取依赖路径地址
func readDependencyPaths() {
	os := strings.TrimSpace(GlobalConfig.Os)
	FfmpegPath, YoutubeDlPath = "ffmpeg", "youtube-dl"
	if os == "" {
		return
	}
	path, err := checkPath("config/ffmpeg/ffmpeg-" + os)
	if err == nil {
		FfmpegPath = path
	}
	path, err = checkPath("config/youtube-dl/youtube-dl-" + os)
	if err == nil {
		YoutubeDlPath = path
	}
}

// 检查路径的文件是否存在
// @param path 要检查的路径
// @return 检测成功的路径
func checkPath(path string) (string, error) {
	validExtensions := []string{"", ".exe", ".sh", ".cmd"}
	for _, ext := range validExtensions {
		newPath := path + ext
		_, err := os.Stat(newPath)
		if err == nil {
			return newPath, nil
		}
	}
	return "", errors.New("找不到依赖可执行文件")
}
