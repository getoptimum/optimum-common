package ratelimit_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/ratelimit"
)

func TestPerSecond_ConcurrentAllowsOverflow(t *testing.T) {
	// Use the in-memory store.
	mem := ratelimit.NewMemoryUsage()

	// Fix "now" and align the window so no reset during the test.
	now := time.Unix(1_700_000_000, 0) // arbitrary fixed timestamp
	_ = mem.WithUsage(func(ratelimit.Usage) (ratelimit.Usage, error) {
		return ratelimit.Usage{
			SecondCount: 0,
			SecondStart: now, // same as "now" -> no rollover in CheckPerSecond
			HourCount:   0,
			HourStart:   now,
			DayBytes:    0,
			DayStart:    now,
		}, nil
	})

	limit := 10
	workers := 200 // Many concurrent calls to amplify the race

	var successes int32
	var wg sync.WaitGroup
	wg.Add(workers)

	// Barrier to start all goroutines at once.
	start := make(chan struct{})

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start // wait for the simultaneous start

			// Every goroutine calls CheckPerSecond with the SAME "now".
			if err := ratelimit.CheckPerSecond(mem, limit, now); err == nil {
				atomic.AddInt32(&successes, 1)
			}
		}()
	}

	// Release the barrier.
	close(start)
	wg.Wait()

	// With correct atomic update semantics, at most "limit" calls should succeed.
	if int(successes) > limit {
		t.Fatalf("per-second limit overflow: got %d successful calls, want <= %d", successes, limit)
	}
}
