package utils_test

import (
	"bytes"
	"encoding/hex"
	"regexp"
	"sort"
	"testing"

	randutil "github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestRandInt64_UniquenessAndSign(t *testing.T) {
	t.Parallel()

	const N = 100_000

	seen := make(map[int64]struct{}, N)
	for i := 0; i < N; i++ {
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

	for i := 0; i < 100_000; i++ {
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

	for i := 0; i < N; i++ {
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

func TestShuffle_PermutationAndNonIdentity(t *testing.T) {
	t.Parallel()

	orig := []int{1, 2, 3, 4, 5, 6, 7}

	// empty and single should not panic and should be stable
	randutil.Shuffle([]int(nil))
	one := []int{42}
	randutil.Shuffle(one)
	require.Equal(t, []int{42}, one)

	// permutation property -> same multiset after shuffle
	a := append([]int(nil), orig...)
	randutil.Shuffle(a)

	sortedA := append([]int(nil), a...)
	sort.Ints(sortedA)
	require.Equal(t, orig, sortedA, "shuffle must be a permutation of the original")

	// with overwhelming probability at least one position differs
	identicalPositions := 0
	for i := range orig {
		if a[i] == orig[i] {
			identicalPositions++
		}
	}
	require.NotEqual(t, len(orig), identicalPositions, "shuffle produced identity order (very unlikely); try rerunning")
}

func TestMsgHash_DeterminismAndOrder(t *testing.T) {
	t.Parallel()

	topic := "topic"
	msg := []byte("hello")

	h1 := randutil.MsgHash(topic, msg)
	h2 := randutil.MsgHash(topic, msg)
	require.Equal(t, h1, h2, "same inputs should yield same hash")

	// ensure order
	h3 := randutil.MsgHash("topic2", msg)
	require.NotEqual(t, h1, h3)

	h4 := randutil.MsgHash(topic, []byte("hello!"))
	require.NotEqual(t, h1, h4)
}

func TestMsgHash_DelimiterAvoidsAmbiguity(t *testing.T) {
	t.Parallel()

	aTopic, aMsg := "ab", []byte("c")
	bTopic, bMsg := "a", []byte("bc")

	ha := randutil.MsgHash(aTopic, aMsg)
	hb := randutil.MsgHash(bTopic, bMsg)
	require.NotEqualf(t, ha, hb, "delimiter-safe hashing should differ, but both were %s", ha)
}

func TestMsgHash_DelimiterAvoidsOtherAmbiguities(t *testing.T) {
	t.Parallel()

	// "abc" + "" vs "ab" + "c"
	ha := randutil.MsgHash("abc", []byte(""))
	hb := randutil.MsgHash("ab", []byte("c"))
	require.NotEqual(t, ha, hb)
}

func TestMsgHash_IsHex_And_Length(t *testing.T) {
	t.Parallel()

	h := randutil.MsgHash("t", []byte("m"))
	require.NotEmpty(t, h)

	// expect hex
	_, err := hex.DecodeString(h)
	require.NoError(t, err, "hash must be hex-encoded")
	require.Equal(t, 64, len(h), "unexpected hash length (expecting SHA-256 hex)")
}

// ========== Benchmarks ==========

var (
	sinkInt64 int64
	sinkInt   int
	sinkStr   string
)

func BenchmarkRandInt64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := randutil.RandInt64()
		if err != nil {
			b.Fatal(err)
		}
		sinkInt64 = v
	}
}

func BenchmarkRandInt64Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var v int64
		for pb.Next() {
			x, err := randutil.RandInt64()
			if err != nil {
				b.Fatal(err)
			}
			v = x
		}
		sinkInt64 = v
	})
}

func BenchmarkRandInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := randutil.RandInt()
		if err != nil {
			b.Fatal(err)
		}
		sinkInt = v
	}
}

func BenchmarkRandBetween(b *testing.B) {
	const minVal, maxVal = 10, 110 // span 100
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := randutil.RandBetween(minVal, maxVal)
		if err != nil {
			b.Fatal(err)
		}
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
			if err != nil {
				b.Fatal(err)
			}
			v = x
		}
		sinkInt = v
	})
}

func BenchmarkMsgHash_Short(b *testing.B) {
	topic := "t"
	msg := []byte("m")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = randutil.MsgHash(topic, msg)
	}
	if sinkStr == "" {
		b.Fatal("empty hash")
	}
}

func BenchmarkMsgHash_Large(b *testing.B) {
	topic := "a-very-long-topic-name-to-avoid-inlining"
	msg := bytes.Repeat([]byte("x"), 64*1024) // 64 KiB
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = randutil.MsgHash(topic, msg)
	}
	if sinkStr == "" {
		b.Fatal("empty hash")
	}
}

func BenchmarkShuffle_IntSlice100(b *testing.B) {
	base := make([]int, 100)
	for i := range base {
		base[i] = i
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cp := append([]int(nil), base...)
		randutil.Shuffle(cp)
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
		randutil.Shuffle(cp)
	}
}

// assert hex ( to catch impl change)
var hexRe = regexp.MustCompile(`\A[0-9a-fA-F]+\z`)

func TestMsgHash_IsHex_WithRegex(t *testing.T) {
	t.Parallel()

	h := randutil.MsgHash("topic", []byte("data"))
	require.Truef(t, hexRe.MatchString(h), "hash is not hex: %q", h)
}
