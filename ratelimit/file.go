package ratelimit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// FileUsage stores Usage backed by a JSON file.
type FileUsage struct {
	mu    sync.Mutex
	path  string
	usage Usage
}

// NewFileUsage creates or loads Usage from the given path.
func NewFileUsage(path string) (*FileUsage, error) {
	fu := &FileUsage{path: path}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &fu.usage)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	now := time.Now()
	if fu.usage.SecondStart.IsZero() {
		fu.usage.SecondStart = now
	}
	if fu.usage.HourStart.IsZero() {
		fu.usage.HourStart = now
	}
	if fu.usage.DayStart.IsZero() {
		fu.usage.DayStart = now
	}
	return fu, nil
}

// GetUsage implements UsageData.
func (f *FileUsage) GetUsage() Usage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.usage
}

// SaveUsage implements UsageData.
func (f *FileUsage) SaveUsage(u Usage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.usage = u
	data, err := json.Marshal(f.usage)
	if err != nil {
		return err
	}
	return os.WriteFile(f.path, data, 0o600)
}
