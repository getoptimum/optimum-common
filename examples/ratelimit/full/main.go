package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/getoptimum/optimum-common/ratelimit"
)

func main() {
	usage := ratelimit.NewMemoryUsage()
	msg := []byte("hello")

	for sent := 0; sent < 5; {
		// Ensure the message itself is within the allowed size.
		if err := ratelimit.CheckMessageSize(int64(len(msg)), 8); err != nil {
			fmt.Println("message rejected:", err)
			break
		}

		now := time.Now()

		// Enforce a per-second rate limit.
		if err := ratelimit.CheckPerSecond(usage, 2, now); err != nil {
			if ratelimit.IsRateLimitError(err) {
				var l *ratelimit.LimitError
				errors.As(err, &l)
				fmt.Printf("per-second limit hit, resets at %s\n", l.ResetAt().Format(time.Kitchen))
				time.Sleep(time.Until(l.ResetAt()))
				continue
			}
			fmt.Println("unexpected error:", err)
			break
		}

		// Enforce a per-hour rate limit.
		if err := ratelimit.CheckPerHour(usage, 3, now); err != nil {
			if ratelimit.IsRateLimitError(err) {
				var l *ratelimit.LimitError
				errors.As(err, &l)
				fmt.Printf("per-hour limit hit, resets at %s\n", l.ResetAt().Format(time.Kitchen))
				break
			}
			fmt.Println("unexpected error:", err)
			break
		}

		// Track bytes sent against a daily quota.
		if err := ratelimit.CheckDaily(usage, int64(len(msg)), 20, now); err != nil {
			if ratelimit.IsRateLimitError(err) {
				var l *ratelimit.LimitError
				errors.As(err, &l)
				fmt.Printf("daily quota reached after %d bytes\n", l.CurrentUsage)
				break
			}
			fmt.Println("unexpected error:", err)
			break
		}

		fmt.Printf("sent message %d\n", sent+1)
		sent++
		time.Sleep(200 * time.Millisecond)
	}
}
