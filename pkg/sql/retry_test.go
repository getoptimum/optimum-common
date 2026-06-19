package sql

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryReadReplica(t *testing.T) {
	t.Parallel()

	calls := 0
	err := retryReadReplica(context.Background(), func() error {
		calls++
		if calls < 2 {
			return &pq.Error{Code: "40001"}
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, calls)

	calls = 0
	sentinel := errors.New("other error")
	err = retryReadReplica(context.Background(), func() error {
		calls++
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, 1, calls)
}
