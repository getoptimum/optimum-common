package syncx_test

import (
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/syncx"
	"github.com/stretchr/testify/require"
)

func TestRWSliceBasicOperations(t *testing.T) {
	t.Parallel()

	rwSlice := syncx.NewRWSlice[int]()
	rwSlice.Add(1)
	rwSlice.Add(2)
	rwSlice.AddBulk([]int{3, 4})

	require.Equal(t, []int{1, 2, 3, 4}, rwSlice.LoadAll())
	require.Equal(t, 4, rwSlice.Len())

	copied := rwSlice.LoadAndErase()
	require.Equal(t, []int{1, 2, 3, 4}, copied)
	require.Zero(t, rwSlice.Len())
	require.Empty(t, rwSlice.LoadAll())
}

func TestRWSliceErase(t *testing.T) {
	t.Parallel()

	rwSlice := syncx.NewRWSlice[string]()
	rwSlice.Add("hello")
	rwSlice.Add("world")
	require.Equal(t, []string{"hello", "world"}, rwSlice.LoadAll())

	rwSlice.Erase()
	require.Zero(t, rwSlice.Len())
	require.Empty(t, rwSlice.LoadAll())
}

func TestRWSliceReplace(t *testing.T) {
	t.Parallel()

	rwSlice := syncx.NewRWSlice[int]()
	rwSlice.AddBulk([]int{1, 2, 3})
	require.Equal(t, []int{1, 2, 3}, rwSlice.LoadAll())

	// Wholesale swap to a different set.
	rwSlice.Replace([]int{42, 1337})
	require.Equal(t, []int{42, 1337}, rwSlice.LoadAll())
	require.Equal(t, 2, rwSlice.Len())

	// Replace with empty is allowed and clears the slice.
	rwSlice.Replace(nil)
	require.Zero(t, rwSlice.Len())
	require.Empty(t, rwSlice.LoadAll())

	// Mutating the input slice after Replace must not affect the stored copy.
	src := []int{7, 8, 9}
	rwSlice.Replace(src)
	src[0] = 99
	require.Equal(t, []int{7, 8, 9}, rwSlice.LoadAll(), "Replace must clone its input")
}

func TestRWSliceConcurrentUsage(t *testing.T) {
	t.Parallel()

	rwSlice := syncx.NewRWSlice[int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rwSlice.Add(i)
		}(i)
	}

	for range 50 {
		wg.Go(func() {
			_ = rwSlice.LoadAll()
		})
	}

	wg.Wait()
	require.Equal(t, 100, rwSlice.Len())
}
