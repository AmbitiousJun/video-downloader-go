// 自定义的令牌桶
package mytokenbucket

import (
	"fmt"
	"sync"
	"time"
	"video-downloader-go/internal/util/mymath"
)

// GlobalBucket 全局的令牌桶, config 包初始化完成后, 会为其赋值
var GlobalBucket *MyTokenBucket

type MyTokenBucket struct {
	capacity          int64        // 桶容量
	tokens            int64        // 当前桶中含有的令牌数
	totalConsume      int64        // 存放总消耗令牌数，用于计算实际下载速率，每秒钟清空一次
	refillRate        int64        // 每秒补充多少令牌（一个令牌表示 1 Byte）
	lastRefillTime    int64        // 最后一次补充令牌的时间
	tokensMutex       sync.Mutex   // 用于控制令牌数的同步读写
	totalConsumeMutex sync.RWMutex // 用于控制总消耗令牌的同步读写
	CurrentRateStr    string       // 记录当前的下载速率
}

const (
	MinConsumeTokens = 1024 // 最少消耗的令牌数
)

// 创建一个令牌桶对象
func NewTokenBucket(refillRate int64) (*MyTokenBucket, error) {
	tb := &MyTokenBucket{}
	tb.capacity = refillRate
	tb.refillRate = refillRate
	tb.tokens = 0
	tb.lastRefillTime = time.Now().UnixMilli()
	tb.totalConsume = 0
	tb.tokensMutex = sync.Mutex{}
	tb.totalConsumeMutex = sync.RWMutex{}
	tb.autoCalcRateLimit()
	return tb, nil
}

// 获取总消耗令牌数
func (tb *MyTokenBucket) TotalConsume() int64 {
	tb.totalConsumeMutex.RLock()
	defer tb.totalConsumeMutex.RUnlock()
	return tb.totalConsume
}

// 下载完成后通知令牌桶，用于计算下载速率
func (tb *MyTokenBucket) CompleteConsume(consume int64) {
	current := tb.TotalConsume()
	tb.setTotalConsume(current + consume)
}

// 消耗一定数量的令牌
// @param request 要消耗的令牌数
// @return 消耗掉的令牌数
func (tb *MyTokenBucket) TryConsume(request int64) int64 {
	tb.tokensMutex.Lock()
	defer tb.tokensMutex.Unlock()

	// 1 补充令牌
	tb.refillTokens()
	if tb.tokens < MinConsumeTokens {
		return 0
	}

	// 2 计算出当前能够消耗的令牌数
	consume := mymath.Min(tb.tokens, request)
	tb.tokens -= consume
	return consume
}

// 补充 token
func (tb *MyTokenBucket) refillTokens() {
	// 1 获取当前时间
	curTime := time.Now().UnixMilli()
	sub := curTime - tb.lastRefillTime
	// 2 补充相应的令牌
	tb.tokens = mymath.Min(tb.capacity, tb.tokens+sub*tb.refillRate/1000)
	// 3 更新时间
	tb.lastRefillTime = curTime
}

// 设置总消耗令牌数
func (tb *MyTokenBucket) setTotalConsume(value int64) {
	tb.totalConsumeMutex.Lock()
	defer tb.totalConsumeMutex.Unlock()
	tb.totalConsume = value
}

// 定时自动计算当前的下载速率
func (tb *MyTokenBucket) autoCalcRateLimit() {
	lastCalcTime := time.Now().UnixMilli()
	tb.totalConsume = tb.TotalConsume()
	var unit int64 = 1024
	var milliUnit int64 = 1000
	lastRateStr := ""
	doCalc := func() {
		currentTime := time.Now().UnixMilli()
		milli := currentTime - lastCalcTime
		// 计算出 MB/s 单位速率
		var rate float64 = 0
		if tb.totalConsume != 0 {
			rate = float64(tb.totalConsume) * float64(milliUnit) / float64(unit) / float64(unit) / float64(milli)
		}
		rateStr := fmt.Sprintf("%.1f", rate)
		if lastRateStr == rateStr {
			return
		}
		tb.CurrentRateStr = fmt.Sprintf("%v MB/s", rateStr)
		lastRateStr = rateStr
		// 清空状态
		tb.setTotalConsume(0)
		lastCalcTime = currentTime
	}
	go func() {
		for {
			doCalc()
			time.Sleep(time.Second)
		}
	}()
}
