package ratelimit

import (
	"errors"
	"fmt"
	"time"
)

// Usage holds counters and reset timestamps for enforcing limits.
type Usage struct {
	SecondCount int       `json:"second_count"`
	SecondStart time.Time `json:"second_start"`
	HourCount   int       `json:"hour_count"`
	HourStart   time.Time `json:"hour_start"`
	DayBytes    int64     `json:"day_bytes"`
	DayStart    time.Time `json:"day_start"`
}

// UsageData abstracts persistence for Usage information.
// Implementations must
// provide atomic read-modify-write semantics via the callback.
type UsageData interface {
	WithUsage(func(Usage) (Usage, error)) error
}

// LimitError represents a general rate limit exceeded error.
type LimitError struct {
	Message      string
	LimitType    string
	CurrentUsage interface{}
	Limit        interface{}
	ResetTime    time.Time
}

func (e *LimitError) Error() string { return e.Message }

// Type returns the category of limit that was exceeded.
func (e *LimitError) Type() string { return e.LimitType }

// ResetAt returns the time when the limit window resets.
func (e *LimitError) ResetAt() time.Time { return e.ResetTime }

func CheckMessageSize(size, max int64) error {
	if size > max {
		return &LimitError{
			Message:      fmt.Sprintf("message size exceeds limit of %d bytes", max),
			LimitType:    "message_size",
			CurrentUsage: size,
			Limit:        max,
		}
	}
	return nil
}

// TODO: fix code duplication below

func CheckPerSecond(data UsageData, limit int, now time.Time) error {
	return data.WithUsage(func(u Usage) (Usage, error) {
		if now.Sub(u.SecondStart) >= time.Second {
			u.SecondStart = now
			u.SecondCount = 0
		}
		if u.SecondCount >= limit {
			return u, &LimitError{
				Message:      fmt.Sprintf("per-second limit reached (%d/sec)", limit),
				LimitType:    "per_second",
				CurrentUsage: u.SecondCount,
				Limit:        limit,
				ResetTime:    u.SecondStart.Add(time.Second),
			}
		}
		u.SecondCount++
		return u, nil
	})
}

func CheckPerHour(data UsageData, limit int, now time.Time) error {
	return data.WithUsage(func(u Usage) (Usage, error) {
		if now.Sub(u.HourStart) >= time.Hour {
			u.HourStart = now
			u.HourCount = 0
		}
		if u.HourCount >= limit {
			return u, &LimitError{
				Message:      fmt.Sprintf("per-hour limit reached (%d/hour)", limit),
				LimitType:    "per_hour",
				CurrentUsage: u.HourCount,
				Limit:        limit,
				ResetTime:    u.HourStart.Add(time.Hour),
			}
		}
		u.HourCount++
		return u, nil
	})
}

func CheckDaily(data UsageData, size, limit int64, now time.Time) error {
	return data.WithUsage(func(u Usage) (Usage, error) {
		if now.Sub(u.DayStart) >= 24*time.Hour {
			u.DayStart = now
			u.DayBytes = 0
		}
		if u.DayBytes+size > limit {
			return u, &LimitError{
				Message:      fmt.Sprintf("daily quota exceeded (%d/%d bytes)", u.DayBytes+size, limit),
				LimitType:    "daily_quota",
				CurrentUsage: u.DayBytes + size,
				Limit:        limit,
				ResetTime:    u.DayStart.Add(24 * time.Hour),
			}
		}
		u.DayBytes += size
		return u, nil
	})
}

// IsRateLimitError reports whether err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var l *LimitError
	return errors.As(err, &l)
}
