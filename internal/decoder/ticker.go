// 控制解析器解析速率

package decoder

import (
	"sync"
	"time"
)

// 可增长的计数器，每次获取数值就按照一定速率增长后返回，到达最大值后恒定不变
type GrowableTicker struct {
	maxVal   int     // 最大值
	growRate float64 // 增长速率
	curGrow  float64 // 当前的增长值
	curVal   int     // 当前的计数值
	isMax    bool    // 是否已经达到最大值
	mu       sync.Mutex
}

// NewGrowableTicker 返回一个可增长的计数器
func NewGrowableTicker(initialVal, maxVal int, growRate float64) *GrowableTicker {
	if initialVal < 0 {
		initialVal = 0
	}
	if maxVal < initialVal {
		maxVal = initialVal
	}

	gt := &GrowableTicker{
		curVal:  initialVal,
		maxVal:  maxVal,
		curGrow: 1,
	}

	// 数不会增长
	if growRate <= 0 {
		gt.isMax = true
	} else {
		gt.growRate = growRate
	}

	return gt
}

// Next 返回下一个计数值，并自动增长
func (gt *GrowableTicker) Next() time.Duration {
	if gt.isMax {
		// 已经到达最大值了，直接返回当前值
		return time.Duration(gt.curVal)
	}

	gt.mu.Lock()
	defer gt.mu.Unlock()

	current := gt.curVal

	// 计算下一个值
	delta := gt.curGrow * gt.growRate
	gt.curGrow += delta
	gt.curVal += int(delta)

	if gt.curVal > gt.maxVal {
		gt.curVal = gt.maxVal
		gt.isMax = true
	}

	return time.Duration(current)
}
