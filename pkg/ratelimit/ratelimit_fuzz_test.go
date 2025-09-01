package ratelimit_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/ratelimit"
)

func FuzzCheckMessageSize(f *testing.F) {
	seeds := []struct{ size, max int64 }{
		{1, 2},
		{5, 5},
		{10, 5},
	}
	for _, s := range seeds {
		f.Add(s.size, s.max)
	}
	f.Fuzz(func(t *testing.T, size int64, max int64) {
		if size < 0 {
			size = -size
		}
		if max < 0 {
			max = -max
		}
		err := ratelimit.CheckMessageSize(size, max)
		if size <= max && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if size > max {
			if err == nil {
				t.Fatalf("expected error for size %d > %d", size, max)
			}
			if !ratelimit.IsRateLimitError(err) {
				t.Fatalf("expected LimitError, got %T", err)
			}
		}
	})
}
