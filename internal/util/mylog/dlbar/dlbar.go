package dlbar

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

// 任务执行状态
const (
	BarStatusError     = iota // 失败
	BarStatusOk               // 成功
	BarStatusExecuting        // 正在执行
)

// 任务执行子状态
const (
	BarChildStatusWaitingDecode = iota // 等待解析
	BarChildStatusDecode               // 正在解析
	BarChildStatusDownload             // 正在下载
	BarChildStatusTransfer             // 正在转换
)

// BarOption 是 Bar 结构的初始化函数
type BarOption func(*Bar)

// Bar 用于记录一个任务的日志
type Bar struct {
	// 为 Bar 对象提供原子更新操作
	Mu sync.Mutex

	// 每一个任务初始化时, 都会被分配到一个全局唯一的 id, 之后以此 id 来更新日志
	Id string

	// 任务当前的状态
	Status int

	// 任务当前的子状态, 只有当状态为正在执行时, 该子状态的值才有效
	ChildStatus int

	// 文件名称
	Name string

	// 提示信息, 当处于不同状态时, 在界面上展示不同的提示信息
	Hint string

	// 百分比, 范围: [0, 100], 只有当子状态为 正在下载 时才有效
	Percent int

	// 文件当前大小(Byte), 只有当子状态为 正在下载 时才有效
	Size int64

	// 日志项列表
	Items []Item
}

func WithStatus(status int) BarOption {
	return func(b *Bar) { b.Status = status }
}

func WithChildStatus(cs int) BarOption {
	return func(b *Bar) { b.ChildStatus = cs }
}

func WithName(name string) BarOption {
	return func(b *Bar) { b.Name = name }
}

func WithHint(hint string) BarOption {
	return func(b *Bar) { b.Hint = hint }
}

func WithPercent(percent int) BarOption {
	return func(b *Bar) { b.Percent = percent }
}

func WithSize(size int64) BarOption {
	return func(b *Bar) { b.Size = size }
}

// NewBar 根据用户提供的选项初始化一个 Bar
func NewBar(options ...BarOption) *Bar {
	b := new(Bar)
	for _, opt := range options {
		opt(b)
	}
	b.Id = uuid.New().String()
	b.Items = []Item{NewStatusItem(), NewHintItem(), NewNameItem()}
	return b
}

// ErrorHint 更新提示信息, 并标记状态为 Error
func (b *Bar) ErrorHint(hint string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Hint = hint
	b.Status = BarStatusError
}

// OkHint 更新提示信息, 并标记为 Ok
func (b *Bar) OkHint(hint string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Hint = hint
	b.Status = BarStatusOk
}

// WaitingDecodeHint 更新提示信息, 并标记为正在等待解析
func (b *Bar) WaitingDecodeHint(hint string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Hint = hint
	b.Status = BarStatusExecuting
	b.ChildStatus = BarChildStatusWaitingDecode
}

// DecodeHint 更新提示信息, 并标记为正在解析
func (b *Bar) DecodeHint(hint string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Hint = hint
	b.Status = BarStatusExecuting
	b.ChildStatus = BarChildStatusDecode
}

// TransferHint 更新提示信息, 并标记为正在转换
func (b *Bar) TransferHint(hint string) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	b.Hint = hint
	b.Status = BarStatusExecuting
	b.ChildStatus = BarChildStatusTransfer
}

// UpdatePercentAndSize 更新百分比和大小, 当且仅当传入值合法时才会更新
func (b *Bar) UpdatePercentAndSize(percent int, size int64) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if percent < 0 || percent > 100 {
		return
	}
	if size < 0 {
		return
	}
	b.Percent = percent
	b.Size = size
	b.Status = BarStatusExecuting
	b.ChildStatus = BarChildStatusDownload
}

// UpdatePercent 更新百分比, 当且仅当传入值合法时才会更新
func (b *Bar) UpdatePercent(percent int) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if percent < 0 || percent > 100 {
		return
	}
	b.Percent = percent
	b.Status = BarStatusExecuting
	b.ChildStatus = BarChildStatusDownload
}

// UpdateSize 更新大小, 当且仅当传入值合法时才会更新
func (b *Bar) UpdateSize(size int64) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if size < 0 {
		return
	}
	b.Size = size
	b.Status = BarStatusExecuting
	b.ChildStatus = BarChildStatusDownload
}

func (b *Bar) String() string {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	sb := strings.Builder{}
	for i := 0; i < len(b.Items); i++ {
		sb.WriteString(b.Items[i].String(b))
		if i < len(b.Items)-1 {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}
