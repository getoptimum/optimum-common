package rlnc

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// BenchmarkKernels compares every kernel this CPU supports on row widths from
// short basis rows to full shards.
func BenchmarkKernels(b *testing.B) {
	rng := rand.New(rand.NewPCG(9, 10))
	src := make([]byte, 8000)
	dst := make([]byte, len(src))
	for index := range src {
		src[index] = byte(rng.Uint32())
		dst[index] = byte(rng.Uint32())
	}

	for _, k := range kernels {
		for _, n := range []int{32, 128, 1024, 8000} {
			b.Run(fmt.Sprintf("kernel=%s/op=xor/n=%d", k, n), func(b *testing.B) {
				withKernel(b, k, func() {
					b.SetBytes(int64(n))
					for b.Loop() {
						xorSlice(dst[:n], src[:n], 0x53)
					}
				})
			})

			b.Run(fmt.Sprintf("kernel=%s/op=mul/n=%d", k, n), func(b *testing.B) {
				withKernel(b, k, func() {
					b.SetBytes(int64(n))
					for b.Loop() {
						mulSlice(dst[:n], src[:n], 0x53)
					}
				})
			})
		}
	}
}
