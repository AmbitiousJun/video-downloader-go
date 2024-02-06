// 用于初始化全局唯一的解析器

package decoder

import (
	"sync"
	"video-downloader-go/internal/config"
	"video-downloader-go/internal/decoder/catcatch"
	"video-downloader-go/internal/decoder/ytdl"
)

type DecoderHolder struct {
	sync.Once // 用于控制同一类型的解析器只初始化一次
	dcd       D
}

var decoderMap map[string]*DecoderHolder

// 初始化 decoderMap
func init() {
	decoderMap = map[string]*DecoderHolder{
		config.DecoderYoutubeDl:  {},
		config.DecoderCatCatchTx: {},
	}
}

// GetDecoder 根据传递的解析器类型返回一个解析器对象
func GetDecoder(use string) D {
	holder, ok := decoderMap[use]
	if !ok {
		return nil
	}

	if holder.dcd != nil {
		return holder.dcd
	}

	if use == config.DecoderYoutubeDl {
		holder.Once.Do(func() {
			holder.dcd = new(ytdl.Decoder)
		})
		return holder.dcd
	}

	if use == config.DecoderCatCatchTx {
		holder.Once.Do(func() {
			holder.dcd = new(catcatch.TxDecoder)
		})
		return holder.dcd
	}

	return nil
}
