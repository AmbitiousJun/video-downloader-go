package util_test

import (
	"fmt"
	"testing"
	"video-downloader-go/internal/util"

	"github.com/pkg/errors"
)

func TestRetryError(t *testing.T) {
	err1 := errors.Wrap(errors.New("网络异常"), "这是一个新的异常")
	if util.IsRetryableError(err1) {
		fmt.Println("识别成功")
	}
}
