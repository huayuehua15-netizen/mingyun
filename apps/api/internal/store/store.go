// Package store 保存排盘结果。
//
// 现在只有内存实现：进程重启即丢，够 M1/M2 前半段联调用。Postgres 实现在
// M2 接入时替换本包的 ChartStore 即可，HTTP 层不需要改。
package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Record 一条盘面记录。
type Record struct {
	ChartID   string
	Method    schema.Method
	Question  string
	Summary   string
	Chart     schema.Chart
	CreatedAt time.Time
	Entropy   *entropy.Toss
}

// ListOptions 列表查询参数。
type ListOptions struct {
	Method string
	Limit  int
	Offset int
}

// ChartStore 盘面存储接口。
type ChartStore interface {
	Save(Record) error
	Get(id string) (Record, bool)
	List(ListOptions) ([]Record, int)
}

// Memory 内存实现，并发安全。
type Memory struct {
	mu    sync.RWMutex
	byID  map[string]Record
	order []string // 创建顺序，自旧到新
}

// NewMemory 构造内存存储。
func NewMemory() *Memory {
	return &Memory{byID: make(map[string]Record)}
}

// Save 写入一条记录。ID 重复时报错，避免覆盖。
func (m *Memory) Save(rec Record) error {
	if rec.ChartID == "" {
		return fmt.Errorf("chart_id 不能为空")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byID[rec.ChartID]; exists {
		return fmt.Errorf("盘面 %s 已存在", rec.ChartID)
	}
	m.byID[rec.ChartID] = rec
	m.order = append(m.order, rec.ChartID)
	return nil
}

// Get 按 ID 取记录。
func (m *Memory) Get(id string) (Record, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.byID[id]
	return rec, ok
}

// List 按创建时间倒序返回记录，并给出过滤后的总数。
func (m *Memory) List(opts ListOptions) ([]Record, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Record, 0, opts.Limit)
	total, skipped := 0, 0
	for i := len(m.order) - 1; i >= 0; i-- {
		rec := m.byID[m.order[i]]
		if opts.Method != "" && string(rec.Method) != opts.Method {
			continue
		}
		total++
		if skipped < opts.Offset {
			skipped++
			continue
		}
		if opts.Limit > 0 && len(out) >= opts.Limit {
			continue
		}
		out = append(out, rec)
	}
	return out, total
}

// Len 返回记录数，供健康检查与测试使用。
func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.byID)
}
