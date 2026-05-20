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
