package syncx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

func TestRoundRobinBalancer_ValuesReturnsACopy(t *testing.T) {
	t.Run("mutating the result does not affect the balancer", func(t *testing.T) {
		// Given a balancer over two values.
		b := syncx.NewRoundRobinBalancer([]string{"a", "b"})

		// When the caller mutates what Values returned.
		got := b.Values()
		got[0] = "mutated"

		// Then the balancer still cycles through its own values.
		require.Equal(t, "a", b.Next())
		require.Equal(t, "b", b.Next())
		require.Equal(t, []string{"a", "b"}, b.Values())
	})

	t.Run("each call returns an independent slice", func(t *testing.T) {
		// Given two reads of the same balancer.
		b := syncx.NewRoundRobinBalancer([]int{1, 2, 3})
		first := b.Values()
		second := b.Values()

		// When one is mutated.
		first[1] = 99

		// Then the other is untouched.
		require.Equal(t, []int{1, 2, 3}, second)
	})
}
