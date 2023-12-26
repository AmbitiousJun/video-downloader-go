package decoder_test

import (
	"log"
	"testing"
	"video-downloader-go/internal/decoder"
)

func TestTicker(t *testing.T) {
	gt := decoder.NewGrowableTicker(10, 5*60, 0.2)
	for i := 0; i < 1000; i++ {
		log.Println(gt.Next())
	}
}
