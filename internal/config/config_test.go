package config_test

import (
	"log"
	"testing"
	"video-downloader-go/internal/config"
)

func load() error {
	return config.Load("../../config/config.yml")
}

func TestRead(t *testing.T) {
	err := load()
	if err != nil {
		t.Error(err)
	}
}

func TestCustoms(t *testing.T) {
	if err := load(); err != nil {
		t.Error(err)
		return
	}

	execFunc := func(dcUrl string) {
		cu := config.G.Decoder.CustomUse(dcUrl)
		ccf := config.G.Decoder.YoutubeDL.CustomCookiesFrom(dcUrl)
		cfc := config.G.Decoder.YoutubeDL.CustomFormatCodes(dcUrl)
		log.Printf("dcUrl: %v, use: %v, cookies-from: %v, format-codes: %v\n", dcUrl, cu, ccf, cfc)
	}

	execFunc("http://example.com/a.mp4")

	execFunc("https://mgtv.com/a.m3u8")
}
