package config

import "errors"

var ffmpegPath string
var youtubeDlPath string

// 全局加载配置
func Load() error {
	ffmpegPath = "1"
	youtubeDlPath = "2"
	return errors.New("初始化失败")
}
