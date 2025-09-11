package sliceutil_test

import (
	"fmt"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils/sliceutil"
	"github.com/stretchr/testify/require"
)

func TestConvertSlice2(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		[]string{"Number: 1", "Number: 2", "Number: 3"},
		sliceutil.MapSlice([]int{1, 2, 3}, func(i int) string {
			return fmt.Sprintf("Number: %d", i)
		}),
	)

	require.Equal(t,
		[]int{2, 4, 6},
		sliceutil.MapSlice([]float32{1, 2, 3}, func(i float32) int {
			return int(i) * 2
		}),
	)

	// Edge cases:
	require.Equal(t,
		[]int{}, // empty slice (not nil)
		sliceutil.MapSlice([]int{}, func(i int) int { return i }),
	)

	var nilSrc []int
	require.Equal(t,
		[]int{},
		sliceutil.MapSlice(nilSrc, func(i int) int { return i }),
	)
}
