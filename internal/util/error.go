package util

const (
	NetworkError = "网络异常" // 网络异常信息，用于区分可重试错误
)

// AnyError 从给定的两个错误中返回任意一个不为空的错误
// 优先返回 err1，如果两个错误都为空，那么返回空
func AnyError(err1, err2 error) error {
	if err1 != nil {
		return err1
	}
	return err2
}
