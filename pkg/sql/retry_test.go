package sql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/sql"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRecoveryConflict(t *testing.T) {
	t.Parallel()
	assert.True(t, sql.IsRecoveryConflict(&pq.Error{Code: "40001"}))
	assert.False(t, sql.IsRecoveryConflict(&pq.Error{Code: "57014"}))
	assert.False(t, sql.IsRecoveryConflict(&pq.Error{Code: "42P01"}))
	assert.False(t, sql.IsRecoveryConflict(errors.New("connection refused")))
}

func TestRetryReadReplica(t *testing.T) {
	t.Parallel()

	t.Run("retries then succeeds", func(t *testing.T) {
		calls := 0
		err := sql.RetryReadReplica(context.Background(), func() error {
			calls++
			if calls < 2 {
				return &pq.Error{Code: "40001"}
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("permanent error no retry", func(t *testing.T) {
		calls := 0
		sentinel := errors.New("syntax error")
		err := sql.RetryReadReplica(context.Background(), func() error {
			calls++
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, 1, calls)
	})

	t.Run("query canceled no retry", func(t *testing.T) {
		calls := 0
		err := sql.RetryReadReplica(context.Background(), func() error {
			calls++
			return &pq.Error{Code: "57014"}
		})
		require.Error(t, err)
		assert.Equal(t, 1, calls)
	})
}
