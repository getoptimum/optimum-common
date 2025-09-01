package ratelimit

import (
	"path/filepath"
	"testing"
	"time"
)

// runSuite executes rate limit tests against the provided UsageData factory.
func runSuite(t *testing.T, factory func() UsageData) {
	t.Run("per-second", func(t *testing.T) {
		data := factory()
		now := time.Now()
		if err := CheckPerSecond(data, 1, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := CheckPerSecond(data, 1, now); err == nil {
			t.Fatalf("expected error for second limit")
		}
		if err := CheckPerSecond(data, 1, now.Add(time.Second)); err != nil {
			t.Fatalf("expected reset after second: %v", err)
		}
	})

	t.Run("per-hour", func(t *testing.T) {
		data := factory()
		now := time.Now()
		if err := CheckPerHour(data, 2, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := CheckPerHour(data, 2, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := CheckPerHour(data, 2, now); err == nil {
			t.Fatalf("expected error for hour limit")
		}
		if err := CheckPerHour(data, 2, now.Add(time.Hour)); err != nil {
			t.Fatalf("expected reset after hour: %v", err)
		}
	})

	t.Run("daily", func(t *testing.T) {
		data := factory()
		now := time.Now()
		if err := CheckDaily(data, 5, 10, now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := CheckDaily(data, 6, 10, now); err == nil {
			t.Fatalf("expected error for quota limit")
		}
		if err := CheckDaily(data, 5, 10, now.Add(24*time.Hour)); err != nil {
			t.Fatalf("expected reset after day: %v", err)
		}
	})

	t.Run("message-size", func(t *testing.T) {
		if err := CheckMessageSize(5, 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := CheckMessageSize(11, 10); err == nil {
			t.Fatalf("expected error for message size")
		}
	})
}

func TestUsageDataImplementations(t *testing.T) {
	runSuite(t, func() UsageData { return NewMemoryUsage() })

	runSuite(t, func() UsageData {
		dir := t.TempDir()
		path := filepath.Join(dir, "usage.json")
		data, err := NewFileUsage(path)
		if err != nil {
			t.Fatalf("file usage init: %v", err)
		}
		return data
	})
}
