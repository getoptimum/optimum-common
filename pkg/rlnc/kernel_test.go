package rlnc

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// guard fills the bytes around a kernel's dst, so a write past either end shows.
const guard = 0xa5

// withKernel runs fn with k as the active kernel.
func withKernel(tb testing.TB, k kernel, fn func()) {
	tb.Helper()

	previous := activeKernel
	activeKernel = k
	defer func() { activeKernel = previous }()

	fn()
}

// referenceMul is the bitwise product, independent of every table a kernel uses.
func referenceMul(dst, src []byte, factor byte, xor bool) {
	for index := range dst {
		if xor {
			dst[index] ^= multiply(factor, src[index])
		} else {
			dst[index] = multiply(factor, src[index])
		}
	}
}

// affine is GF2P8AFFINEQB with a zero constant, as the Intel SDM defines it:
// output bit i is the parity of matrix byte 7-i ANDed with the input.
func affine(matrix uint64, value byte) byte {
	var output byte
	for bit := range 8 {
		row := byte(matrix >> (8 * (7 - bit)))
		parity := 0
		for selected := row & value; selected != 0; selected &= selected - 1 {
			parity ^= 1
		}

		output |= byte(parity) << bit
	}

	return output
}

func TestGFNIMatricesMatchMultiply(t *testing.T) {
	for factor := range 256 {
		for value := range 256 {
			require.Equal(t, multiply(byte(factor), byte(value)), affine(gfniMatrices[factor], byte(value)),
				"factor %d value %d", factor, value)
		}
	}
}

func TestNibbleTablesMatchMultiply(t *testing.T) {
	for factor := range 256 {
		tables := &nibbleTables[factor]
		for value := range 256 {
			product := tables[value&15] ^ tables[16+value>>4]
			require.Equal(t, multiply(byte(factor), byte(value)), product, "factor %d value %d", factor, value)
		}
	}
}

func TestSupportedKernels(t *testing.T) {
	require.NotEmpty(t, kernels)
	require.Equal(t, kernelGeneric, kernels[len(kernels)-1])
	require.Equal(t, kernels[0], activeKernel)

	for _, k := range kernels {
		require.NotEqual(t, "unknown", k.String())
	}
	require.Equal(t, "unknown", kernel(255).String())

	t.Logf("kernels on this CPU: %v", kernels)
}

// kernelLengths covers every remainder of the 16, 32 and 64-byte vector steps,
// the masked tail, and the shard sizes the audit uses.
func kernelLengths() []int {
	lengths := make([]int, 0, 140)
	for n := range 130 {
		lengths = append(lengths, n)
	}

	return append(lengths, 191, 255, 256, 257, 1000, 4095, 4096, 8000)
}

// checkKernel runs one operation through the active kernel on dst and src at
// the given offsets into guarded buffers, and compares it with referenceMul.
func checkKernel(t *testing.T, rng *rand.Rand, n, dstOffset, srcOffset int, factor byte, xor bool) {
	t.Helper()

	dstBuf := make([]byte, dstOffset+n+64)
	for index := range dstBuf {
		dstBuf[index] = guard
	}
	dst := dstBuf[dstOffset : dstOffset+n]
	for index := range dst {
		dst[index] = byte(rng.Uint32())
	}

	srcBuf := make([]byte, srcOffset+n+64)
	for index := range srcBuf {
		srcBuf[index] = byte(rng.Uint32())
	}
	src := srcBuf[srcOffset : srcOffset+n]
	srcBefore := slices.Clone(srcBuf)

	want := slices.Clone(dst)
	referenceMul(want, src, factor, xor)

	if xor {
		xorSlice(dst, src, factor)
	} else {
		mulSlice(dst, src, factor)
	}

	require.Equal(t, want, dst, "n %d factor %d xor %t", n, factor, xor)
	require.Equal(t, srcBefore, srcBuf, "src changed")
	for index, value := range dstBuf {
		if index < dstOffset || index >= dstOffset+n {
			require.Equal(t, byte(guard), value, "write outside dst at %d, n %d", index, n)
		}
	}
}

func TestKernelsMatchReference(t *testing.T) {
	for _, k := range kernels {
		t.Run(k.String(), func(t *testing.T) {
			withKernel(t, k, func() {
				rng := rand.New(rand.NewPCG(1, uint64(k)))
				for _, n := range kernelLengths() {
					for _, xor := range []bool{false, true} {
						// Unaligned starts, and every factor once per length.
						checkKernel(t, rng, n, int(rng.Uint32()%64), int(rng.Uint32()%64), byte(n), xor)
						checkKernel(t, rng, n, 0, 0, byte(rng.Uint32()), xor)
					}
				}

				for factor := range 256 {
					checkKernel(t, rng, 100, 3, 7, byte(factor), false)
					checkKernel(t, rng, 100, 3, 7, byte(factor), true)
				}
			})
		})
	}
}

// Every kernel keeps the index-order semantics of the bitwise loop when dst
// and src share a backing array.
func TestKernelsHandleOverlap(t *testing.T) {
	testCases := []struct {
		name               string
		dstStart, srcStart int
	}{
		{name: "exact", dstStart: 0, srcStart: 0},
		{name: "src ahead of dst", dstStart: 0, srcStart: 5},
		{name: "dst ahead of src", dstStart: 5, srcStart: 0},
		{name: "dst one vector ahead", dstStart: 64, srcStart: 0},
	}

	const n = 200
	for _, k := range kernels {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s/%s", k, tc.name), func(t *testing.T) {
				withKernel(t, k, func() {
					rng := rand.New(rand.NewPCG(3, 4))
					buf := make([]byte, n+64)
					for index := range buf {
						buf[index] = byte(rng.Uint32())
					}

					want := slices.Clone(buf)
					referenceMul(want[tc.dstStart:tc.dstStart+n], want[tc.srcStart:tc.srcStart+n], 0x53, true)
					XorScaled(buf[tc.dstStart:tc.dstStart+n], buf[tc.srcStart:tc.srcStart+n], 0x53)
					require.Equal(t, want, buf)

					if tc.dstStart == tc.srcStart {
						scaled := slices.Clone(buf[:n])
						referenceMul(scaled, scaled, 0x53, false)
						Scale(buf[:n], 0x53)
						require.Equal(t, scaled, buf[:n])
					}
				})
			})
		}
	}
}

func TestOverlapsBehind(t *testing.T) {
	buf := make([]byte, 8)

	require.False(t, overlapsBehind(buf, buf))
	require.False(t, overlapsBehind(buf[:4], buf[2:]))
	require.True(t, overlapsBehind(buf[2:], buf[:4]))
	require.False(t, overlapsBehind(buf[4:], buf[:4]), "adjacent slices do not overlap")
	require.False(t, overlapsBehind(buf, nil))
	require.False(t, overlapsBehind(buf, make([]byte, 8)))
}

// FuzzKernels checks every kernel this CPU supports against referenceMul.
func FuzzKernels(f *testing.F) {
	f.Add([]byte("0123456789abcdef0123456789abcdef!"), []byte("fedcba9876543210fedcba9876543210?"), byte(0x53), uint8(1))
	f.Add(make([]byte, 64), make([]byte, 64), byte(0xff), uint8(0))

	f.Fuzz(func(t *testing.T, dst, src []byte, factor byte, offset uint8) {
		start := int(offset) % (len(src) + 1)
		src = src[start:]
		if len(src) < len(dst) {
			dst = dst[:len(src)]
		}

		want := slices.Clone(dst)
		referenceMul(want, src, factor, true)
		scaled := slices.Clone(dst)
		referenceMul(scaled, src, factor, false)

		for _, k := range kernels {
			withKernel(t, k, func() {
				got := slices.Clone(dst)
				xorSlice(got, src, factor)
				require.Equal(t, want, got, "kernel %s xor", k)

				got = slices.Clone(dst)
				mulSlice(got, src, factor)
				require.Equal(t, scaled, got, "kernel %s mul", k)
			})
		}
	})
}
