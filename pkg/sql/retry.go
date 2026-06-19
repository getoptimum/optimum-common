package sql

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
)

const (
	readReplicaRetryMax     = 3
	readReplicaRetryBackoff = 100 * time.Millisecond
)

func isRecoveryConflict(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code == "40001"
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
		case <-time.After(readReplicaRetryBackoff * time.Duration(attempt+1)):
		}
	}
	return err
}
