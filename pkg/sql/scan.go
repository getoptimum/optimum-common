package sql

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/sqlscan"
)

func QueryRowsToStruct[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]*T, error) {
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	res := make([]*T, 0, 100)
	for rows.Next() {
		var t T
		if errS := sqlscan.NewRowScanner(rows).Scan(&t); errS != nil {
			return nil, errS
		}
		res = append(res, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func QueryRowToStruct[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) (*T, error) {
	var t T
	if err := sqlscan.Get(ctx, conn, &t, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get row: %w", err)
	}
	return &t, nil
}

func QueryRowsPrimitive[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]T, error) {
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	res := make([]T, 0, 64)
	for rows.Next() {
		var v T
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, rows.Err()
}

// QueryRowsToStructWithRetry is QueryRowsToStruct wrapped for read-replica conflict retry.
func QueryRowsToStructWithRetry[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]*T, error) {
	var res []*T
	err := retryReadReplica(ctx, func() error {
		var err error
		res, err = QueryRowsToStruct[T](ctx, conn, query, args...)
		return err
	})
	return res, err
}

// QueryRowsPrimitiveWithRetry is QueryRowsPrimitive wrapped for read-replica conflict retry.
func QueryRowsPrimitiveWithRetry[T any](ctx context.Context, conn sqlscan.Querier, query string, args ...any) ([]T, error) {
	var res []T
	err := retryReadReplica(ctx, func() error {
		var err error
		res, err = QueryRowsPrimitive[T](ctx, conn, query, args...)
		return err
	})
	return res, err
}
