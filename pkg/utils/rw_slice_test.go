package utils_test

import (
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestRWSliceBasicOperations(t *testing.T) {
	t.Parallel()

	rwSlice := utils.NewRWSlice[int]()
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

	rwSlice := utils.NewRWSlice[string]()
	rwSlice.Add("hello")
	rwSlice.Add("world")
	require.Equal(t, []string{"hello", "world"}, rwSlice.LoadAll())

	rwSlice.Erase()
	require.Zero(t, rwSlice.Len())
	require.Empty(t, rwSlice.LoadAll())
}

func TestRWSliceConcurrentUsage(t *testing.T) {
	t.Parallel()

	rwSlice := utils.NewRWSlice[int]()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rwSlice.Add(i)
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rwSlice.LoadAll()
		}()
	}

	wg.Wait()
	require.Equal(t, 100, rwSlice.Len())
}
