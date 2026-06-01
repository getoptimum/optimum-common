package sql

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
)

const (
	ReadReplicaRetryMax     = 3
	ReadReplicaRetryBackoff = 100 * time.Millisecond
)

// IsRecoveryConflict reports Postgres 40001 hot-standby conflicts worth retrying on read-replica queries.
func IsRecoveryConflict(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code == "40001"
}

// RetryReadReplica runs fn, retrying on 40001 read-replica recovery conflicts.
func RetryReadReplica(ctx context.Context, fn func() error) error {
	var err error
	for attempt := range ReadReplicaRetryMax + 1 {
		if err = fn(); err == nil || !IsRecoveryConflict(err) {
			return err
		}
		if ctx.Err() != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(ReadReplicaRetryBackoff * time.Duration(attempt+1)):
		}
	}
	return err
}
