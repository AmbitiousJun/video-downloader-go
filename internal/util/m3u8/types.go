package m3u8

// ts 文件信息
type TsMeta struct {

	// HeadUrl 该字段一开始是为了兼容 EXT-X-MAP 而设置,
	// 由于找到更简单的兼容方式, 故现在该字段已弃用
	HeadUrl string
	Url     string // 真实请求 url
	Index   int    // 记录 ts 文件是位于第几个，便于后期合成
}
