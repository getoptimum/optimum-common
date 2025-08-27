package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/getoptimum/optimum-common/ratelimit"
)

func main() {
	fmt.Println("Message size check:")
	if err := ratelimit.CheckMessageSize(int64(len("hello world")), 8); err != nil {
		fmt.Printf("  %v\n\n", err)
	}
	fmt.Println("Per-second limit:")
	usage := ratelimit.NewMemoryUsage()
	for i := 1; i <= 5; i++ {
		err := ratelimit.CheckPerSecond(usage, 2, time.Now())
		if err != nil {
			var lerr *ratelimit.LimitError
			if errors.As(err, &lerr) {
				wait := time.Until(lerr.ResetAt()).Round(time.Millisecond)
				fmt.Printf("  attempt %d blocked: %s (wait %v)\n", i, lerr.Message, wait)
				time.Sleep(wait)
				continue
			}
			fmt.Printf("  unexpected error: %v\n", err)
			return
		}
		fmt.Printf("  attempt %d allowed\n", i)
	}
}
