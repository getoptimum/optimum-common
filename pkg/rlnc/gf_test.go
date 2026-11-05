package rlnc_test

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/rlnc"
)

func TestGFMultiply(t *testing.T) {
	testCases := []struct {
		name        string
		left, right byte
		output      byte
	}{
		{name: "zero", left: 0x53, right: 0, output: 0},
		{name: "identity", left: 0x53, right: 1, output: 0x53},
		{name: "no reduction", left: 2, right: 3, output: 6},
		{name: "reduces by 0x11d", left: 2, right: 0x80, output: 0x1d},
		{name: "max", left: 0xff, right: 0xff, output: 0xe2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.output, rlnc.GFMultiply(tc.left, tc.right))
			require.Equal(t, tc.output, rlnc.GFMultiply(tc.right, tc.left))
		})
	}
}

func TestGFInverse(t *testing.T) {
	require.Zero(t, rlnc.GFInverse(0))

	for value := 1; value < 256; value++ {
		require.Equal(t, byte(1), rlnc.GFMultiply(byte(value), rlnc.GFInverse(byte(value))), "value %d", value)
	}
}

func TestScale(t *testing.T) {
	row := []byte{0, 1, 3, 0x80}
	rlnc.Scale(row, 2)
	require.Equal(t, []byte{0, 2, 6, 0x1d}, row)
}

func TestXorScaled(t *testing.T) {
	dst := []byte{1, 2, 3}
	rlnc.XorScaled(dst, []byte{1, 1, 0x80}, 2)
	require.Equal(t, []byte{3, 0, 3 ^ 0x1d}, dst)
}

func TestLeadingColumn(t *testing.T) {
	require.Equal(t, -1, rlnc.LeadingColumn(nil))
	require.Equal(t, -1, rlnc.LeadingColumn([]byte{0, 0}))
	require.Equal(t, 0, rlnc.LeadingColumn([]byte{5, 0}))
	require.Equal(t, 2, rlnc.LeadingColumn([]byte{0, 0, 7}))
}

func TestAuditSymbol(t *testing.T) {
	shards := [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	coefficients := []byte{2, 0, 0x80}
	data := combine(coefficients, shards)

	testCases := []struct {
		name         string
		coefficients []byte
		data         []byte
		shards       [][]byte
		valid        bool
	}{
		{name: "coded symbol", coefficients: coefficients, data: data, shards: shards, valid: true},
		{name: "source symbol", coefficients: []byte{0, 1, 0}, data: shards[1], shards: shards, valid: true},
		{name: "unit coefficients", coefficients: []byte{1, 1, 1}, data: combine([]byte{1, 1, 1}, shards), shards: shards, valid: true},
		{name: "zero coefficients and zero data", coefficients: []byte{0, 0, 0}, data: []byte{0, 0, 0}, shards: shards, valid: true},
		{name: "single shard generation", coefficients: []byte{7}, data: combine([]byte{7}, shards[:1]), shards: shards[:1], valid: true},
		{name: "first byte polluted", coefficients: coefficients, data: withByte(data, 0), shards: shards},
		{name: "last byte polluted", coefficients: coefficients, data: withByte(data, 2), shards: shards},
		{name: "source symbol polluted", coefficients: []byte{0, 1, 0}, data: withByte(shards[1], 1), shards: shards},
		{name: "zero coefficients with nonzero data", coefficients: []byte{0, 0, 0}, data: []byte{0, 1, 0}, shards: shards},
		{name: "polluted coefficients", coefficients: []byte{2, 1, 0x80}, data: data, shards: shards},
		{name: "too few coefficients", coefficients: coefficients[:2], data: data, shards: shards},
		{name: "too many coefficients", coefficients: append([]byte{0}, coefficients...), data: data, shards: shards},
		{name: "short data", coefficients: coefficients, data: data[:2], shards: shards},
		{name: "long data", coefficients: coefficients, data: append(slices.Clone(data), 0), shards: shards},
		{name: "uneven shards", coefficients: coefficients, data: data, shards: [][]byte{{1, 2, 3}, {4, 5}, {7, 8, 9}}},
		{name: "nil shard", coefficients: coefficients, data: data, shards: [][]byte{{1, 2, 3}, nil, {7, 8, 9}}},
		{name: "empty data", coefficients: []byte{1}, shards: [][]byte{{}}},
		{name: "no coefficients", data: data, shards: shards},
		{name: "no shards"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.valid, rlnc.AuditSymbol(tc.coefficients, tc.data, tc.shards))
		})
	}
}

// TestAuditSymbolMatchesOracle checks AuditSymbol against the bitwise oracle
// over the generation sizes the config allows (k up to 128, shards up to 8000
// bytes), shard sizes on both sides of the audit's block boundary, and
// coefficient rows with zeros and ones.
func TestAuditSymbolMatchesOracle(t *testing.T) {
	const blockSize = 4096

	rng := rand.New(rand.NewPCG(1, 2))
	generations := []struct{ k, s int }{
		{k: 1, s: 1},
		{k: 2, s: 2},
		{k: 16, s: 7},
		{k: 64, s: blockSize - 1},
		{k: 16, s: blockSize},
		{k: 2, s: blockSize + 1},
		{k: 1, s: 2*blockSize + 3},
		{k: 128, s: 8000},
	}

	for _, generation := range generations {
		k, s := generation.k, generation.s

		shards := randomShards(rng, k, s)
		coefficients := randomCoefficients(rng, k)
		data := combine(coefficients, shards)

		shardsBefore := cloneShards(shards)
		coefficientsBefore := slices.Clone(coefficients)
		dataBefore := slices.Clone(data)

		require.True(t, rlnc.AuditSymbol(coefficients, data, shards), "k=%d s=%d", k, s)

		require.Equal(t, shardsBefore, shards)
		require.Equal(t, coefficientsBefore, coefficients)
		require.Equal(t, dataBefore, data)

		for _, index := range []int{0, s - 1, min(blockSize-1, s-1), min(blockSize, s-1), rng.IntN(s)} {
			require.False(t, rlnc.AuditSymbol(coefficients, withByte(data, index), shards), "k=%d s=%d byte %d", k, s, index)
		}

		// Changing coefficient row adds a multiple of shards[row], which shows
		// in data unless that shard is all zeros.
		row := rng.IntN(k)
		if slices.ContainsFunc(shards[row], func(value byte) bool { return value != 0 }) {
			pollutedCoefficients := slices.Clone(coefficients)
			pollutedCoefficients[row] ^= 1
			require.False(t, rlnc.AuditSymbol(pollutedCoefficients, data, shards), "k=%d s=%d row %d", k, s, row)
		}
	}
}

func BenchmarkAuditSymbol(b *testing.B) {
	for _, bc := range []struct{ k, s int }{{k: 16, s: 4096}, {k: 64, s: 8000}, {k: 128, s: 8000}} {
		rng := rand.New(rand.NewPCG(3, 4))
		shards := randomShards(rng, bc.k, bc.s)

		coded := randomBytes(rng, bc.k)
		source := make([]byte, bc.k)
		source[bc.k/2] = 1

		for _, symbol := range []struct {
			name         string
			coefficients []byte
		}{{name: "coded", coefficients: coded}, {name: "source", coefficients: source}} {
			data := combine(symbol.coefficients, shards)

			b.Run(fmt.Sprintf("k=%d/s=%d/%s", bc.k, bc.s, symbol.name), func(b *testing.B) {
				b.SetBytes(int64(bc.k * bc.s))
				b.ReportAllocs()
				for b.Loop() {
					if !rlnc.AuditSymbol(symbol.coefficients, data, shards) {
						b.Fatal("valid symbol failed the audit")
					}
				}
			})
		}
	}
}

// combine computes Σⱼ coefficientsⱼ·shardsⱼ with the bitwise GFMultiply, as an
// oracle independent of AuditSymbol's lookup table.
func combine(coefficients []byte, shards [][]byte) []byte {
	data := make([]byte, len(shards[0]))
	for row, coefficient := range coefficients {
		rlnc.XorScaled(data, shards[row], coefficient)
	}

	return data
}

func withByte(data []byte, index int) []byte {
	polluted := slices.Clone(data)
	polluted[index] ^= 1

	return polluted
}

func randomShards(rng *rand.Rand, k, s int) [][]byte {
	shards := make([][]byte, k)
	for row := range shards {
		shards[row] = randomBytes(rng, s)
	}

	return shards
}

// randomCoefficients draws a row where about a quarter of the entries are 0 and
// a quarter are 1, the two values AuditSymbol handles apart.
func randomCoefficients(rng *rand.Rand, k int) []byte {
	coefficients := randomBytes(rng, k)
	for i := range coefficients {
		switch rng.IntN(4) {
		case 0:
			coefficients[i] = 0
		case 1:
			coefficients[i] = 1
		}
	}

	return coefficients
}

func cloneShards(shards [][]byte) [][]byte {
	clones := make([][]byte, len(shards))
	for i, shard := range shards {
		clones[i] = slices.Clone(shard)
	}

	return clones
}

func randomBytes(rng *rand.Rand, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(rng.Uint32())
	}

	return b
}
