package m3u8

// ts 文件信息
type TsMeta struct {
	// 真实请求 url
	Url string
	// 记录 ts 文件是位于第几个，便于后期合成
	Index int
}
