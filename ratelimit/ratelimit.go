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

// UsageData to abstract persistence for Usage information.
type UsageData interface {
	GetUsage() Usage
	SaveUsage(Usage) error
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

func CheckPerSecond(data UsageData, limit int, now time.Time) error {
	usage := data.GetUsage()
	if now.Sub(usage.SecondStart) >= time.Second {
		usage.SecondStart = now
		usage.SecondCount = 0
	}
	if usage.SecondCount >= limit {
		return &LimitError{
			Message:      fmt.Sprintf("per-second limit reached (%d/sec)", limit),
			LimitType:    "per_second",
			CurrentUsage: usage.SecondCount,
			Limit:        limit,
			ResetTime:    usage.SecondStart.Add(time.Second),
		}
	}
	usage.SecondCount++
	return data.SaveUsage(usage)
}

func CheckPerHour(data UsageData, limit int, now time.Time) error {
	usage := data.GetUsage()
	if now.Sub(usage.HourStart) >= time.Hour {
		usage.HourStart = now
		usage.HourCount = 0
	}
	if usage.HourCount >= limit {
		return &LimitError{
			Message:      fmt.Sprintf("per-hour limit reached (%d/hour)", limit),
			LimitType:    "per_hour",
			CurrentUsage: usage.HourCount,
			Limit:        limit,
			ResetTime:    usage.HourStart.Add(time.Hour),
		}
	}
	usage.HourCount++
	return data.SaveUsage(usage)
}

func CheckDaily(data UsageData, size, limit int64, now time.Time) error {
	usage := data.GetUsage()
	if now.Sub(usage.DayStart) >= 24*time.Hour {
		usage.DayStart = now
		usage.DayBytes = 0
	}
	if usage.DayBytes+size > limit {
		return &LimitError{
			Message:      fmt.Sprintf("daily quota exceeded (%d/%d bytes)", usage.DayBytes+size, limit),
			LimitType:    "daily_quota",
			CurrentUsage: usage.DayBytes + size,
			Limit:        limit,
			ResetTime:    usage.DayStart.Add(24 * time.Hour),
		}
	}
	usage.DayBytes += size
	return data.SaveUsage(usage)
}

// IsRateLimitError reports whether err is a RateLimitError.
func IsRateLimitError(err error) bool {
	var l *LimitError
	return errors.As(err, &l)
}
