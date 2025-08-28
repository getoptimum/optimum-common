// demonstrates usage of the ratelimit package
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
		if err := ratelimit.CheckMessageSize(int64(len(msg)), 8); err != nil {
			fmt.Println("message rejected:", err)
			break
		}

		if err := ratelimit.CheckPerSecond(usage, 2, time.Now()); err != nil {
			var l *ratelimit.LimitError
			errors.As(err, &l)
			fmt.Printf("per-second limit hit, resets at %s\n", l.ResetAt().Format(time.Kitchen))
			time.Sleep(time.Until(l.ResetAt()))
			continue
		}

		if err := ratelimit.CheckDaily(usage, int64(len(msg)), 10, time.Now()); err != nil {
			var l *ratelimit.LimitError
			errors.As(err, &l)
			fmt.Printf("daily quota reached after %d bytes\n", l.CurrentUsage)
			break
		}

		fmt.Printf("sent message %d\n", sent+1)
		sent++
		time.Sleep(200 * time.Millisecond)
	}
}
