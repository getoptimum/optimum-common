package ratelimit

import (
	"sync"
	"time"
)

// MemoryUsage stores Usage in memory.
type MemoryUsage struct {
	mu    sync.Mutex
	usage Usage
}

// NewMemoryUsage returns a MemoryUsage with zeroed counters.
func NewMemoryUsage() *MemoryUsage {
	now := time.Now()
	return &MemoryUsage{usage: Usage{SecondStart: now, HourStart: now, DayStart: now}}
}

// GetUsage implements UsageData.
func (m *MemoryUsage) GetUsage() Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usage
}

// SaveUsage implements UsageData.
func (m *MemoryUsage) SaveUsage(u Usage) error {
	m.mu.Lock()
	m.usage = u
	m.mu.Unlock()
	return nil
}
