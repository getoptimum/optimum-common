package rand_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	randutil "github.com/getoptimum/optimum-common/pkg/rand"
)

var (
	sinkInt64 int64
	sinkInt   int
)

func BenchmarkRandInt64(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
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
	for b.Loop() {
		v, err := randutil.RandInt()
		require.NoError(b, err)
		sinkInt = v
	}
}

func BenchmarkRandBetween(b *testing.B) {
	const minVal, maxVal = 10, 110 // span 100
	b.ReportAllocs()
	for b.Loop() {
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
	for b.Loop() {
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
	for b.Loop() {
		cp := append([]int(nil), base...)
		require.NoError(b, randutil.Shuffle(cp))
	}
}
