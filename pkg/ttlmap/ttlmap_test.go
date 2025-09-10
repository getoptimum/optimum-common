package ttlmap_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/ttlmap"
	"github.com/stretchr/testify/require"
)

func TestTTLMap(t *testing.T) {
	t.Run("basic put get and expiry", func(t *testing.T) {
		maxTTL := time.Second
		cleanup := time.Second
		m := ttlmap.NewTTLMap[string, int](maxTTL, cleanup)

		m.Put("key", 42)
		require.Equal(t, 1, m.Len())

		val, ok := m.Get("key")
		require.True(t, ok)
		require.Equal(t, 42, val)

		require.Eventually(t, func() bool {
			_, ok = m.Get("key")
			return !ok
		}, 5*cleanup, cleanup)
	})

	t.Run("concurrent access", func(t *testing.T) {
		maxTTL := 2 * time.Second
		cleanup := time.Second
		m := ttlmap.NewTTLMap[string, int](maxTTL, cleanup)
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Put("key"+strconv.Itoa(i), i)
			}(i)
		}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if v, ok := m.Get("key" + strconv.Itoa(i)); ok {
					require.Equal(t, i, v)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("do method", func(t *testing.T) {
		m := ttlmap.NewTTLMap[string, int](time.Minute, time.Minute)
		m.Put("a", 1)
		executed := false
		ok := m.Do("a", func(v int) { executed = true; require.Equal(t, 1, v) })
		require.True(t, ok)
		require.True(t, executed)

		ok = m.Do("missing", func(v int) {})
		require.False(t, ok)
	})
}
