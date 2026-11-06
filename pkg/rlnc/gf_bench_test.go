package rlnc_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/rlnc"
)

func BenchmarkXorScaled(b *testing.B) {
	for _, size := range []int{128, 8000} {
		rng := rand.New(rand.NewPCG(5, 6))
		dst, src := randomBytes(rng, size), randomBytes(rng, size)

		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for b.Loop() {
				rlnc.XorScaled(dst, src, 0x53)
			}
		})
	}
}

func BenchmarkScale(b *testing.B) {
	row := randomBytes(rand.New(rand.NewPCG(7, 8)), 256)

	b.SetBytes(int64(len(row)))
	for b.Loop() {
		rlnc.Scale(row, 0x53)
	}
}

func BenchmarkGFInverse(b *testing.B) {
	for b.Loop() {
		for value := 1; value < 256; value++ {
			rlnc.GFInverse(byte(value))
		}
	}
}
