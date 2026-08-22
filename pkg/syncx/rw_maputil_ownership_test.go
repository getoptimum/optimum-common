package syncx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

func TestRWMap_OwnsItsMap(t *testing.T) {
	t.Run("NewRWMapFromStdMap copies the source", func(t *testing.T) {
		src := map[string]int{"a": 1}
		m := syncx.NewRWMapFromStdMap(src)

		src["a"] = 99
		src["b"] = 2

		v, ok := m.Load("a")
		require.True(t, ok)
		require.Equal(t, 1, v)
		require.Equal(t, 1, m.Len())
	})

	t.Run("Replace copies the replacement", func(t *testing.T) {
		m := syncx.NewRWMap[string, int]()
		m.Store("old", 1)

		src := map[string]int{"a": 1}
		previous := m.Replace(src)
		require.Equal(t, map[string]int{"old": 1}, previous)

		src["a"] = 99
		src["b"] = 2

		v, ok := m.Load("a")
		require.True(t, ok)
		require.Equal(t, 1, v)
		require.Equal(t, 1, m.Len())
	})
}
