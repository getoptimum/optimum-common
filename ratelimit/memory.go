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

func (m *MemoryUsage) WithUsage(fn func(Usage) (Usage, error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, err := fn(m.usage)
	m.usage = u
	return err
}

func (m *MemoryUsage) SaveUsage(u Usage) error {
	m.mu.Lock()
	m.usage = u
	m.mu.Unlock()
	return nil
}
