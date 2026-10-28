package rlnc_test

import (
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
