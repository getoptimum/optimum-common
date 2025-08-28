// Concurrent demo of all rate-limit checks.
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/getoptimum/optimum-common/ratelimit"
)

func main() {
	fmt.Println("=== SCENARIO 1: Burst + backoff on per-second (high hour/daily caps) ===")
	{
		usage := ratelimit.NewMemoryUsage()
		cfg := Limits{MaxMsgSize: 16, PerSecond: 3, PerHour: 10_000, DailyBytes: 1 << 30}
		runScenarioBurstBackoff(usage, cfg, 6, 3*time.Second)
	}

	fmt.Println("\n=== SCENARIO 2: Hit per-hour quickly (very small hourly cap) ===")
	{
		usage := ratelimit.NewMemoryUsage()
		cfg := Limits{MaxMsgSize: 32, PerSecond: 1000, PerHour: 8, DailyBytes: 1 << 30}
		runScenarioHitPerHour(usage, cfg, 4)
	}

	fmt.Println("\n=== SCENARIO 3: Hit daily quota (small daily bytes) ===")
	{
		usage := ratelimit.NewMemoryUsage()
		cfg := Limits{MaxMsgSize: 64, PerSecond: 1000, PerHour: 1000, DailyBytes: 200} // ~200 bytes
		runScenarioHitDaily(usage, cfg, 5)
	}
}

/* ------------------------------- Shared ---------------------------------- */

type Limits struct {
	MaxMsgSize int64
	PerSecond  int
	PerHour    int
	DailyBytes int64
}

type sendResult struct {
	worker int
	n      int    // message sequence for that worker
	err    error  // nil on success
	info   string // optional note
}

func limiterChecks(usage ratelimit.UsageData, cfg Limits, msg []byte, now time.Time) error {
	// 1) Size
	if err := ratelimit.CheckMessageSize(int64(len(msg)), cfg.MaxMsgSize); err != nil {
		return err
	}
	// 2) Per-second
	if err := ratelimit.CheckPerSecond(usage, cfg.PerSecond, now); err != nil {
		return err
	}
	// 3) Per-hour
	if err := ratelimit.CheckPerHour(usage, cfg.PerHour, now); err != nil {
		return err
	}
	// 4) Daily-bytes
	if err := ratelimit.CheckDaily(usage, int64(len(msg)), cfg.DailyBytes, now); err != nil {
		return err
	}
	return nil
}

func describeErr(prefix string, err error) string {
	if err == nil {
		return prefix + "OK"
	}
	if ratelimit.IsRateLimitError(err) {
		var le *ratelimit.LimitError
		if errors.As(err, &le) && le != nil {
			// We assume LimitError exposes reset time and current usage stats.
			return fmt.Sprintf("%sLIMIT: %v | resets at %s | current=%d limit=%d",
				prefix, le, le.ResetAt().Format(time.Kitchen), le.CurrentUsage, le.Limit)
		}
		return fmt.Sprintf("%sLIMIT: %v", prefix, err)
	}
	return fmt.Sprintf("%sERROR: %v", prefix, err)
}

/* -------------------------- Scenario 1: Backoff --------------------------- */

func runScenarioBurstBackoff(usage ratelimit.UsageData, cfg Limits, workers int, duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan sendResult, 1024)

	// One worker will purposely send an oversize message once to show CheckMessageSize.
	oversizeOnce := make(chan struct{}, 1)
	oversizeOnce <- struct{}{} // token to use exactly once

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			// Per-goroutine RNG (not shared). Seed mixed with id for variety.
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

			defer wg.Done()
			n := 0
			baseMsg := []byte("hello")
			t := time.NewTicker(60 * time.Millisecond)
			defer t.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					n++

					// 1 in 20 chance to trigger message-size error, but only once globally:
					msg := baseMsg
					select {
					case <-oversizeOnce:
						// Make it 2x over the limit
						msg = make([]byte, cfg.MaxMsgSize*2)
						for i := range msg {
							msg[i] = 'X'
						}
					default:
						// jitter message length a tiny bit
						if r.Intn(20) == 0 {
							msg = append([]byte{}, baseMsg...)
							msg = append(msg, '!')
						}
					}

					now := time.Now()
					err := limiterChecks(usage, cfg, msg, now)
					if err != nil && ratelimit.IsRateLimitError(err) {
						// Backoff only for per-second (in this scenario we set hour/daily high).
						var le *ratelimit.LimitError
						if errors.As(err, &le) && le != nil {
							backoff := time.Until(le.ResetAt())
							if backoff > 0 && backoff < time.Second {
								time.Sleep(backoff)
							}
						}
					}
					results <- sendResult{worker: id, n: n, err: err}
				}
			}
		}(w)
	}

	// Collector
	go func() {
		wg.Wait()
		close(results)
	}()

	success := 0
	failSize := 0
	failSecond := 0

	for r := range results {
		switch {
		case r.err == nil:
			success++
			if success%10 == 0 {
				fmt.Printf("[burst] sent ok total=%d\n", success)
			}
		case ratelimit.IsRateLimitError(r.err):
			var le *ratelimit.LimitError
			errors.As(r.err, &le)
			// We distinguish per-second by type/content; assuming the error exposes a name or limit window via fmt.Stringer or fields.
			fmt.Println(describeErr(fmt.Sprintf("[burst w%02d m%03d] ", r.worker, r.n), r.err))
			failSecond++
		default:
			// Only other error we expect here is CheckMessageSize
			fmt.Println(describeErr(fmt.Sprintf("[burst w%02d m%03d] ", r.worker, r.n), r.err))
			failSize++
		}
	}

	fmt.Printf("[burst] summary: ok=%d size_errs=%d per_sec_hits=%d\n", success, failSize, failSecond)
}

/* ------------------------- Scenario 2: Per-hour hit ----------------------- */

func runScenarioHitPerHour(usage ratelimit.UsageData, cfg Limits, workers int) {
	var wg sync.WaitGroup
	results := make(chan sendResult, 1024)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := []byte("hour")
			for n := 1; ; n++ {
				now := time.Now()
				err := limiterChecks(usage, cfg, msg, now)
				results <- sendResult{worker: id, n: n, err: err}
				if err != nil {
					// Stop this worker once we hit the per-hour wall
					return
				}
				// tiny stagger so we trip per-second rarely
				time.Sleep(30 * time.Millisecond)
			}
		}(w)
	}

	// Collector
	go func() {
		wg.Wait()
		close(results)
	}()

	ok := 0
	var hourErrs int
	for r := range results {
		if r.err == nil {
			ok++
			continue
		}
		if ratelimit.IsRateLimitError(r.err) {
			fmt.Println(describeErr(fmt.Sprintf("[hour  w%02d m%03d] ", r.worker, r.n), r.err))
			hourErrs++
		} else {
			fmt.Println(describeErr(fmt.Sprintf("[hour  w%02d m%03d] ", r.worker, r.n), r.err))
		}
	}
	fmt.Printf("[hour] summary: ok=%d per_hour_hits=%d (cap=%d)\n", ok, hourErrs, cfg.PerHour)
}

/* ------------------------- Scenario 3: Daily bytes ------------------------ */

func runScenarioHitDaily(usage ratelimit.UsageData, cfg Limits, workers int) {
	var wg sync.WaitGroup
	results := make(chan sendResult, 1024)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			// Per-goroutine RNG.
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

			defer wg.Done()
			// Randomize payload sizes a bit to see 'CurrentUsage' climb
			for n := 1; ; n++ {
				msg := make([]byte, 10+r.Intn(30)) // up to ~40B
				for i := range msg {
					msg[i] = 'D'
				}
				now := time.Now()
				err := limiterChecks(usage, cfg, msg, now)
				results <- sendResult{worker: id, n: n, err: err, info: fmt.Sprintf("bytes=%d", len(msg))}
				if err != nil {
					return // stop after daily cap
				}
				time.Sleep(20 * time.Millisecond)
			}
		}(w)
	}

	// Collector
	go func() {
		wg.Wait()
		close(results)
	}()

	ok := 0
	var dailyHits int
	for r := range results {
		if r.err == nil {
			ok++
			continue
		}
		if ratelimit.IsRateLimitError(r.err) {
			fmt.Println(describeErr(fmt.Sprintf("[daily w%02d m%03d %s] ", r.worker, r.n, r.info), r.err))
			dailyHits++
		} else {
			fmt.Println(describeErr(fmt.Sprintf("[daily w%02d m%03d %s] ", r.worker, r.n, r.info), r.err))
		}
	}
	fmt.Printf("[daily] summary: ok=%d daily_hits=%d (bytes cap=%d)\n", ok, dailyHits, cfg.DailyBytes)
}
