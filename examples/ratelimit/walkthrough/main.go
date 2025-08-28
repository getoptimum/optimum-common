package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/getoptimum/optimum-common/ratelimit"
)

func main() {
	// Store usage counters in memory. For persistence across restarts,
	// see ratelimit.NewFileUsage.
	usage := ratelimit.NewMemoryUsage()

	messages := []string{"hi", "hello", "world", "!!"}

	for i, msg := range messages {
		// enforce a maximum message size of 5 bytes
		if err := ratelimit.CheckMessageSize(int64(len(msg)), 5); err != nil {
			fmt.Println("size check:", err)
			continue
		}

		now := time.Now()
		// allow at most 2 messages per second
		if err := ratelimit.CheckPerSecond(usage, 2, now); err != nil {
			var lerr *ratelimit.LimitError
			errors.As(err, &lerr)
			fmt.Printf("per-second limit hit on message %d: %v\n", i+1, lerr)
			// wait for the window to reset before retrying
			time.Sleep(time.Until(lerr.ResetAt()))
			continue
		}

		// enforce a daily quota of 10 bytes total
		if err := ratelimit.CheckDaily(usage, int64(len(msg)), 10, now); err != nil {
			fmt.Println("daily quota:", err)
			break
		}

		fmt.Println("sent:", msg)
		time.Sleep(200 * time.Millisecond)
	}
}
