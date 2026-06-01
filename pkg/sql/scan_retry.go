package sql

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"
)

// QueryRowsToStructWithRetry is QueryRowsToStruct wrapped for read-replica conflict retry.
func QueryRowsToStructWithRetry[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]*T, error) {
	var res []*T
	err := RetryReadReplica(ctx, func() error {
		var err error
		res, err = QueryRowsToStruct[T](ctx, conn, query, args...)
		return err
	})
	return res, err
}

// QueryRowsPrimitiveWithRetry is QueryRowsPrimitive wrapped for read-replica conflict retry.
func QueryRowsPrimitiveWithRetry[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]T, error) {
	var res []T
	err := RetryReadReplica(ctx, func() error {
		var err error
		res, err = QueryRowsPrimitive[T](ctx, conn, query, args...)
		return err
	})
	return res, err
}
