package m3u8

// ts 文件信息
type TsMeta struct {
	// 视频头部 url
	// 当 m3u8 文件中出现 "#EXT-X-MAP" 前缀时，就需要在下载每个 ts 时额外再下载一个视频头部，下载完成后进行拼接
	HeadUrl string
	Url     string // 真实请求 url
	Index   int    // 记录 ts 文件是位于第几个，便于后期合成
}
