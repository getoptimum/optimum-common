package sql

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/getoptimum/optimum-common/pkg/rand"
)

// A standby cancels reads for as long as it takes to replay the conflicting WAL, so the old
// linear 100/200/300ms never escaped a vacuum stall. The cap must stay under the caller's
// poll interval: measurements' metrics_loop is 30s and its gitops alert saturates at ~25s,
// so this schedule's ~16s of waiting is already close to being that alert. Do not grow it.
const (
	readReplicaRetryMax        = 6 // attempts after the first; 7 calls total
	readReplicaRetryBase       = 250 * time.Millisecond
	readReplicaRetryMaxBackoff = 8 * time.Second
	readReplicaRetryJitter     = 0.2 // +/-20%, so concurrent retriers do not wake in lockstep
)

// 40001 is also serialization_failure under SSI on a primary, where a retry may need to
// re-derive state rather than repeat the call. Every caller of these helpers reads ReadDB();
// keep it that way and do not wrap primary writes with them.
func isRecoveryConflict(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code == "40001"
}

// readReplicaBackoff returns the delay before the attempt after a failed one.
// A var so tests can substitute a fast schedule.
var readReplicaBackoff = func(attempt int) time.Duration {
	d := min(readReplicaRetryBase<<attempt, readReplicaRetryMaxBackoff)
	spread := int(float64(d) * readReplicaRetryJitter)
	if spread <= 0 {
		return d
	}
	// Losing jitter is not worth failing the query over, so a bad rand source falls back
	// to the unjittered delay.
	offset, err := rand.RandBetween(-spread, spread)
	if err != nil {
		return d
	}
	return d + time.Duration(offset)
}

// retryReadReplica re-runs fn while it fails with a recovery conflict, waiting out the
// backoff schedule between attempts and returning the last error once they are spent.
func retryReadReplica(ctx context.Context, fn func() error) error {
	var err error
	for attempt := range readReplicaRetryMax + 1 {
		if err = fn(); err == nil || !isRecoveryConflict(err) {
			return err
		}
		// Return on the last attempt rather than sleeping first: the loop would otherwise
		// wait the capped ~8s with nothing left to retry, so a persistent conflict came
		// back that much later than the schedule claims.
		if attempt == readReplicaRetryMax {
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
