// 存放整个应用的上下文信息
package appctx

import (
	"context"
	"sync"
)

// 上下文对象
var ctx context.Context

// 上下文终止函数
var cancel context.CancelFunc

// 多协程同步对象
var wg sync.WaitGroup

// 初始化应用上下文
func init() {
	ctx, cancel = context.WithCancel(context.Background())
	wg = sync.WaitGroup{}
}

// 获取协程同步对象
func WaitGroup() *sync.WaitGroup {
	return &wg
}

// 获取终止函数
func CancelFunc() context.CancelFunc {
	return cancel
}

// 获取上下文
func Context() context.Context {
	return ctx
}

// 批量完成任务
func BatchDone(count int) {
	for i := 0; i < count; i++ {
		wg.Done()
	}
}
