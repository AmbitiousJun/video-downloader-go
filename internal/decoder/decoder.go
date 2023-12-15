package decoder

// 解析器通用接口
type D interface {
	// 解析并获取下载链接列表
	// 对于 youtube-dl 解析器，有可能会返回两条链接，因为部分站点的解析结果是音视频分开的
	// 其他解析器通常只返回一条链接
	FetchDownloadLinks(url string) ([]string, error)
}
