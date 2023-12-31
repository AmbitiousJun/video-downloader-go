// 定义用于命令行下载器的任务列表类型

package meta

import "sync"

// 任务列表双端队列结构，泛型取值为 Video 和 Download
type TaskDeque[T any] struct {
	list []*T       // 存储数据的切片
	mu   sync.Mutex // 协程同步
}

// Range 用于遍历队列元素
func (td *TaskDeque[T]) Range(itemHandler func(item *T, index int)) {
	for i := 0; i < td.Size(); i++ {
		itemHandler(td.Get(i), i)
	}
}

// OfferLast 从队尾添加一个元素
func (td *TaskDeque[T]) OfferLast(val *T) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.list = append(td.list, val)
}

// OfferFirst 从队首添加一个元素
func (td *TaskDeque[T]) OfferFirst(val *T) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.list = append([]*T{val}, td.list...)
}

// Get 返回队列指定索引的元素，越界返回空
func (td *TaskDeque[T]) Get(index int) *T {
	if index < 0 || index >= td.Size() {
		return nil
	}
	return td.list[index]
}

// Size 返回队列的大小
func (td *TaskDeque[T]) Size() int {
	return len(td.list)
}

// Empty 返回队列是否为空
func (td *TaskDeque[T]) Empty() bool {
	return td.Size() == 0
}

// PollFirst 返回队首元素，如果队列为空，返回 nil
func (td *TaskDeque[T]) PollFirst() *T {
	td.mu.Lock()
	defer td.mu.Unlock()
	if td.Empty() {
		return nil
	}
	val := td.list[0]
	td.list = td.list[1:]
	return val
}

// PollLast 返回队尾元素，如果队列为空，返回 nil
func (td *TaskDeque[T]) PollLast() *T {
	td.mu.Lock()
	defer td.mu.Unlock()
	if td.Empty() {
		return nil
	}
	val := td.list[td.Size()-1]
	td.list = td.list[:td.Size()-1]
	return val
}
