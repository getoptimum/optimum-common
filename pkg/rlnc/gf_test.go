package rlnc_test

import (
	"fmt"
	"math/rand/v2"
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

// TestScaledMatchesMultiply checks the table-driven row operations against
// GFMultiply for every factor, including the 0 and 1 shortcuts.
func TestScaledMatchesMultiply(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	src := randomBytes(rng, 300)
	base := randomBytes(rng, len(src))

	for factor := range 256 {
		scaled := append([]byte(nil), src...)
		rlnc.Scale(scaled, byte(factor))

		sum := append([]byte(nil), base...)
		rlnc.XorScaled(sum, src, byte(factor))

		for index, value := range src {
			product := rlnc.GFMultiply(byte(factor), value)
			require.Equal(t, product, scaled[index], "Scale factor %d index %d", factor, index)
			require.Equal(t, base[index]^product, sum[index], "XorScaled factor %d index %d", factor, index)
		}
	}
}

func TestXorScaledReadsOnlyLenDst(t *testing.T) {
	for _, factor := range []byte{0, 1, 2} {
		dst := []byte{1, 2}
		rlnc.XorScaled(dst, []byte{1, 1, 0xff}, factor)
		require.Len(t, dst, 2)
	}
}

// TestXorScaledOverlap pins the in-order semantics of the original bitwise
// loop when dst and src share a backing array, for every factor path.
func TestXorScaledOverlap(t *testing.T) {
	reference := func(dst, src []byte, factor byte) {
		for index := range dst {
			dst[index] ^= rlnc.GFMultiply(factor, src[index])
		}
	}

	testCases := []struct {
		name             string
		dstStart, dstEnd int
		srcStart, srcEnd int
	}{
		{name: "src ahead of dst", dstStart: 0, dstEnd: 3, srcStart: 1, srcEnd: 4},
		{name: "dst ahead of src", dstStart: 1, dstEnd: 4, srcStart: 0, srcEnd: 3},
		{name: "exact", dstStart: 0, dstEnd: 4, srcStart: 0, srcEnd: 4},
	}

	for _, tc := range testCases {
		for _, factor := range []byte{0, 1, 2, 0x53} {
			t.Run(fmt.Sprintf("%s/factor=%d", tc.name, factor), func(t *testing.T) {
				got := []byte{1, 2, 3, 4}
				want := []byte{1, 2, 3, 4}

				rlnc.XorScaled(got[tc.dstStart:tc.dstEnd], got[tc.srcStart:tc.srcEnd], factor)
				reference(want[tc.dstStart:tc.dstEnd], want[tc.srcStart:tc.srcEnd], factor)

				require.Equal(t, want, got)
			})
		}
	}
}

func TestLeadingColumn(t *testing.T) {
	require.Equal(t, -1, rlnc.LeadingColumn(nil))
	require.Equal(t, -1, rlnc.LeadingColumn([]byte{0, 0}))
	require.Equal(t, 0, rlnc.LeadingColumn([]byte{5, 0}))
	require.Equal(t, 2, rlnc.LeadingColumn([]byte{0, 0, 7}))
}
