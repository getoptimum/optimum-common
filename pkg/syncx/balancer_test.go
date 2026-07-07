package syncx_test

import (
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/syncx"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinBalancer(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		ints := []int{1, 2, 3}
		balancer := syncx.NewRoundRobinBalancer(ints)

		for i := range 10 {
			expected := ints[i%len(ints)]
			require.Equal(t, expected, balancer.Next())
		}

		require.Equal(t, ints, balancer.Values())
	})

	t.Run("strings", func(t *testing.T) {
		strs := []string{"a", "b", "c"}
		balancerStrings := syncx.NewRoundRobinBalancer(strs)

		for i := range 10 {
			expected := strs[i%len(strs)]
			require.Equal(t, expected, balancerStrings.Next())
		}

		require.Equal(t, strs, balancerStrings.Values())
	})

	t.Run("concurrent", func(t *testing.T) {
		var (
			wg            sync.WaitGroup
			mu            sync.Mutex
			numGoroutines = 10
			numIterations = 1000
			ints          = []int{1, 2, 3, 4, 5}
		)
		balancer := syncx.NewRoundRobinBalancer(ints)
		results := make([]int, numGoroutines*numIterations)

		for i := range numGoroutines {
			wg.Add(1)
			go func(startIdx int) {
				defer wg.Done()
				for j := range numIterations {
					result := balancer.Next()
					mu.Lock()
					results[startIdx*numIterations+j] = result
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		expectedCount := make(map[int]int)
		for _, v := range ints {
			expectedCount[v] = numGoroutines * numIterations / len(ints)
		}

		actualCount := make(map[int]int)
		for _, v := range results {
			actualCount[v]++
		}

		require.Equal(t, expectedCount, actualCount)
	})
}
