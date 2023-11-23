// 定义一个队列结构，用于全局控制日志
package log

import "sync"

// 日志结构
type logItem struct {
	string
}

// 存放日志的队列
type logQueue struct {
	items []*logItem
}

var mu sync.RWMutex

// 获取队列的长度
func (lq *logQueue) Len() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(lq.items)
}

// 从队头中取出一个元素
func (lq *logQueue) pollFirst() *logItem {
	if lq.Len() == 0 {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	item := lq.items[0]
	lq.items = lq.items[1:]
	return item
}

// 往队列的末尾添加一个元素
func (lq *logQueue) offerLast(item *logItem) {
	mu.Lock()
	defer mu.Unlock()
	lq.items = append(lq.items, item)
}
