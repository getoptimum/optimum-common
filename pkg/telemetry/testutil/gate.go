package testutil

import (
	"sync/atomic"
	"testing"
)

func WithBool(t *testing.T, b *atomic.Bool, v bool) {
	t.Helper()
	prev := b.Load()
	b.Store(v)
	t.Cleanup(func() { b.Store(prev) })
}
