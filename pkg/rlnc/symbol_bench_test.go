package rlnc_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/rlnc"
)

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
