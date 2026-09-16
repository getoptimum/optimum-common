package sql

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/lib/pq"
)

// Backoff schedule for read-replica recovery conflicts (40001).
//
// A hot-standby conflict is not a brief blip: the standby cancels queries for as long as it
// takes to replay the conflicting WAL, which for a vacuum pass over a large table is minutes,
// not milliseconds. The previous schedule (3 retries, linear 100/200/300ms) spent its whole
// ~600ms budget inside the same stall and so never recovered -- every conflict still surfaced
// as an error. Backing off exponentially to ~24s total gives the standby time to catch up.
//
// The ceiling is deliberately below a typical caller's poll interval so a retrying query does
// not routinely outlive the tick that issued it.
const (
	readReplicaRetryMax        = 6 // attempts after the first; 7 calls total
	readReplicaRetryBase       = 250 * time.Millisecond
	readReplicaRetryMaxBackoff = 8 * time.Second
	readReplicaRetryJitter     = 0.2 // +/-20%, so concurrent retriers do not wake in lockstep
)

func isRecoveryConflict(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code == "40001"
}

// readReplicaBackoff returns the delay before the attempt following a failed one.
// Exponential from readReplicaRetryBase, capped, then jittered.
// It is a var so tests can substitute a fast schedule.
var readReplicaBackoff = func(attempt int) time.Duration {
	d := readReplicaRetryBase << attempt // 250ms, 500ms, 1s, 2s, 4s, 8s, ...
	if d > readReplicaRetryMaxBackoff || d <= 0 {
		d = readReplicaRetryMaxBackoff
	}
	spread := float64(d) * readReplicaRetryJitter
	return time.Duration(float64(d) - spread + rand.Float64()*2*spread) //nolint:gosec // G404: jitter, not security
}

func retryReadReplica(ctx context.Context, fn func() error) error {
	var err error
	for attempt := range readReplicaRetryMax + 1 {
		if err = fn(); err == nil || !isRecoveryConflict(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(readReplicaBackoff(attempt)):
		}
	}
	return err
}
