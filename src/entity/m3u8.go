// 处理 m3u8 文件的相关实体
package entity

// ts 文件信息
type TsMeta struct {
	// 真实请求 url
	Url string
	// 记录 ts 文件是位于第几个，便于后期合成
	Index int
}

// 创建一个新的 TsMeta 结构
func NewTsMeta(url string, index int) *TsMeta {
	return &TsMeta{Url: url, Index: index}
}
