package rand_test

import (
	cryptorand "crypto/rand"
	"sort"
	"strings"
	"testing"

	randutil "github.com/getoptimum/optimum-common/pkg/rand"
	"github.com/stretchr/testify/require"
)

func TestRandInt64_UniquenessAndSign(t *testing.T) {
	t.Parallel()

	const N = 100_000

	seen := make(map[int64]struct{}, N)
	for i := range N {
		v, err := randutil.RandInt64()
		require.NoError(t, err)

		// must be non-negative
		require.GreaterOrEqual(t, v, int64(0))

		_, dup := seen[v]
		require.Falsef(t, dup, "unexpected duplicate at i=%d: %d", i, v)
		seen[v] = struct{}{}
	}
}

func TestRandInt_NonNegative(t *testing.T) {
	t.Parallel()

	for range 100_000 {
		v, err := randutil.RandInt()
		require.NoError(t, err)
		require.GreaterOrEqual(t, v, 0)
	}
}

func TestRandBetween_InvalidRange(t *testing.T) {
	t.Parallel()

	_, err := randutil.RandBetween(5, 5)
	require.Error(t, err)

	_, err = randutil.RandBetween(10, 5)
	require.Error(t, err)
}

func TestRandBetween_BoundsAndCoverage(t *testing.T) {
	t.Parallel()

	const (
		minVal = 10
		maxVal = 20 // span = 10
		span   = maxVal - minVal
		N      = 200_000
	)
	buckets := make([]int, span)

	for range N {
		v, err := randutil.RandBetween(minVal, maxVal)
		require.NoError(t, err)
		require.GreaterOrEqual(t, v, minVal)
		require.Less(t, v, maxVal)
		buckets[v-minVal]++
	}

	// every bucket hit at least once with high probability
	for i, c := range buckets {
		require.NotZerof(t, c, "bucket %d (value %d) never hit", i, i+minVal)
	}
}

func TestShuffle_Permutation(t *testing.T) {
	t.Parallel()

	orig := []int{1, 2, 3, 4, 5, 6, 7}

	require.NoError(t, randutil.Shuffle([]int(nil)))
	one := []int{42}
	require.NoError(t, randutil.Shuffle(one))
	require.Equal(t, []int{42}, one)

	a := append([]int(nil), orig...)
	require.NoError(t, randutil.Shuffle(a))

	sortedA := append([]int(nil), a...)
	sort.Ints(sortedA)
	require.Equal(t, orig, sortedA, "shuffle must be a permutation of the original")
}

func TestShuffle_PropagatesRandomnessError(t *testing.T) {
	reader := cryptorand.Reader
	cryptorand.Reader = strings.NewReader("")
	t.Cleanup(func() {
		cryptorand.Reader = reader
	})

	values := []int{1, 2}
	require.Error(t, randutil.Shuffle(values))
	require.Equal(t, []int{1, 2}, values)
}
