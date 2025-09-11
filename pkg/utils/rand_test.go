package utils_test

import (
	"testing"

	randutil "github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestRandInt64(t *testing.T) {
	matches := map[int64]struct{}{}
	for i := 0; i < 1_000_000; i++ {
		val := randutil.RandInt64()
		_, ok := matches[val]
		require.False(t, ok)
		matches[val] = struct{}{}
	}
}

func TestRandBetween(t *testing.T) {
	minimum := 10
	maximum := 20
	for i := 0; i < 1_000_000; i++ {
		val := randutil.RandBetween(minimum, maximum)
		require.True(t, val >= minimum && val < maximum, val)
	}
}
