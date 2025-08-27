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

// WithUsage implements UsageData.WithUsage.
func (f *FileUsage) WithUsage(fn func(Usage) (Usage, error)) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, err := fn(f.usage)
	f.usage = u
	data, werr := json.Marshal(f.usage)
	if werr != nil {
		return werr
	}
	if werr = os.WriteFile(f.path, data, 0o600); werr != nil {
		return werr
	}
	return err
}
