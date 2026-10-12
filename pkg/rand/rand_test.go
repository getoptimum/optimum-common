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

var (
	sinkInt64 int64
	sinkInt   int
)

func BenchmarkRandInt64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := randutil.RandInt64()
		require.NoError(b, err)
		sinkInt64 = v
	}
}

func BenchmarkRandInt64Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var v int64
		for pb.Next() {
			x, err := randutil.RandInt64()
			require.NoError(b, err)
			v = x
		}
		sinkInt64 = v
	})
}

func BenchmarkRandInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := randutil.RandInt()
		require.NoError(b, err)
		sinkInt = v
	}
}

func BenchmarkRandBetween(b *testing.B) {
	const minVal, maxVal = 10, 110 // span 100
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := randutil.RandBetween(minVal, maxVal)
		require.NoError(b, err)
		sinkInt = v
	}
}

func BenchmarkRandBetweenParallel(b *testing.B) {
	const minVal, maxVal = 10, 1_010
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var v int
		for pb.Next() {
			x, err := randutil.RandBetween(minVal, maxVal)
			require.NoError(b, err)
			v = x
		}
		sinkInt = v
	})
}

func BenchmarkShuffle_IntSlice100(b *testing.B) {
	base := make([]int, 100)
	for i := range base {
		base[i] = i
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cp := append([]int(nil), base...)
		require.NoError(b, randutil.Shuffle(cp))
	}
}

func BenchmarkShuffle_IntSlice1e4(b *testing.B) {
	base := make([]int, 10_000)
	for i := range base {
		base[i] = i
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cp := append([]int(nil), base...)
		require.NoError(b, randutil.Shuffle(cp))
	}
}
