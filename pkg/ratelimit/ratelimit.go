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

func CheckMessageSize(size, limit int64) error {
	if size > limit {
		return &LimitError{
			Message:      fmt.Sprintf("message size exceeds limit of %d bytes", limit),
			LimitType:    "message_size",
			CurrentUsage: size,
			Limit:        limit,
		}
	}
	return nil
}

func checkWindow(
	data UsageData,
	limit int,
	now time.Time,
	startPtr func(*Usage) *time.Time,
	countPtr func(*Usage) *int,
	window time.Duration,
	limitType, msgFmt string,
) error {
	return data.WithUsage(func(u Usage) (Usage, error) {
		start := startPtr(&u)
		count := countPtr(&u)
		if now.Sub(*start) >= window {
			*start = now
			*count = 0
		}
		if *count >= limit {
			return u, &LimitError{
				Message:      fmt.Sprintf(msgFmt, limit),
				LimitType:    limitType,
				CurrentUsage: *count,
				Limit:        limit,
				ResetTime:    start.Add(window),
			}
		}
		*count++
		return u, nil
	})
}

func CheckPerSecond(data UsageData, limit int, now time.Time) error {
	return checkWindow(
		data, limit, now,
		func(u *Usage) *time.Time { return &u.SecondStart },
		func(u *Usage) *int { return &u.SecondCount },
		time.Second,
		"per_second",
		"per-second limit reached (%d/sec)",
	)
}

func CheckPerHour(data UsageData, limit int, now time.Time) error {
	return checkWindow(
		data, limit, now,
		func(u *Usage) *time.Time { return &u.HourStart },
		func(u *Usage) *int { return &u.HourCount },
		time.Hour,
		"per_hour",
		"per-hour limit reached (%d/hour)",
	)
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
