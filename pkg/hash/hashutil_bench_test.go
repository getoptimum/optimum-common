package hash_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/hash"
)

func BenchmarkHashXXHash(b *testing.B) {
	data := []byte("Hello, World!")
	for b.Loop() {
		hash.XXHash(data)
	}
}

func BenchmarkHashSHA256String(b *testing.B) {
	data := []byte("Hello, World!")
	for b.Loop() {
		hash.SHA256String(data)
	}
}
