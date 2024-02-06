// 切片操作工具包
package myslice

// Filter 对给定切片进行过滤, 返回过滤完成后的新切片
func Filter[T any](src []T, filterFunc func(elm T) bool) []T {
	res := make([]T, 0)
	for _, elm := range src {
		if filterFunc(elm) {
			res = append(res, elm)
		}
	}
	return res
}
