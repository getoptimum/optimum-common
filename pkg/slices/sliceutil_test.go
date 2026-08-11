package slices_test

import (
	"fmt"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/slices"
	"github.com/stretchr/testify/require"
)

type testCase[T any] struct {
	input      T
	inputSlice []T
	expected   T
	want       bool
}

type stringer string

func (s stringer) String() string {
	return string(s)
}

func TestConvertSlice(t *testing.T) {
	t.Parallel()

	require.Equal(t,
		[]string{"Number: 1", "Number: 2", "Number: 3"},
		slices.MapSlice([]int{1, 2, 3}, func(i int) string {
			return fmt.Sprintf("Number: %d", i)
		}),
	)

	require.Equal(t,
		[]int{2, 4, 6},
		slices.MapSlice([]float32{1, 2, 3}, func(i float32) int {
			return int(i) * 2
		}),
	)

	// Edge cases:
	require.Equal(t,
		[]int{}, // empty slice (not nil)
		slices.MapSlice([]int{}, func(i int) int { return i }),
	)

	var nilSrc []int
	require.Equal(t,
		[]int{},
		slices.MapSlice(nilSrc, func(i int) int { return i }),
	)
}

func TestUniqueComparable(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{}, slices.UniqueComparable([]string(nil)))
	require.Equal(t, []string{}, slices.UniqueComparable([]string{}))
	require.Equal(t, []string{"a", "b", "c"}, slices.UniqueComparable([]string{"a", "b", "c"}))
	require.Equal(t, []string{"a", "b", "c"}, slices.UniqueComparable([]string{"a", "b", "a", "c", "b"}))
	require.Equal(t, []string{"a"}, slices.UniqueComparable([]string{"a", "a", "a"}))
	require.Equal(t, []string{"", "a", "b"}, slices.UniqueComparable([]string{"", "a", "", "b", "a"}))
	require.Equal(t, []string{"a", "A"}, slices.UniqueComparable([]string{"a", "A", "a"}))
	require.Equal(t, []string{"c", "a", "b", "d"}, slices.UniqueComparable([]string{"c", "a", "b", "a", "c", "d"}))
}

func TestUniqueSlice(t *testing.T) {
	t.Parallel()

	table := []testCase[[]stringer]{
		{
			input:    []stringer{"a", "b", "a", "c", "b", "d", "a"},
			expected: []stringer{"a", "b", "c", "d"},
		},
		{
			input:    []stringer{"A", "a", "A", "a"},
			expected: []stringer{"A", "a"},
		},
		{
			input:    []stringer{"", "", "a", "", "a"},
			expected: []stringer{"", "a"},
		},
		{
			input:    []stringer{"a", " a", "a ", "a\t", "a"},
			expected: []stringer{"a", " a", "a ", "a\t"},
		},
		{
			input:    []stringer{"é", "e\u0301", "é"},
			expected: []stringer{"é", "e\u0301"},
		},
		{
			input:    []stringer{"x", "x", "x", "y", "z", "y", "z"},
			expected: []stringer{"x", "y", "z"},
		},
		{
			input:    []stringer{"k", "l", "m", "n"},
			expected: []stringer{"k", "l", "m", "n"},
		},
		{
			input:    []stringer{},
			expected: []stringer{},
		},
	}
	for _, item := range table {
		require.Equal(t, item.expected, slices.UniqueSlice[stringer](item.input))
	}
}

func TestContainsInSlice(t *testing.T) {
	t.Parallel()
	var nilSlice []int
	cases := []testCase[int]{
		{inputSlice: []int{1, 2, 3, 4}, input: 3, want: true},
		{inputSlice: []int{9, 8, 7}, input: 9, want: true},
		{inputSlice: []int{9, 8, 7}, input: 7, want: true},
		{inputSlice: []int{1, 2, 3}, input: 4, want: false},
		{inputSlice: []int{}, input: 1, want: false},
		{inputSlice: nilSlice, input: 1, want: false},
		{inputSlice: []int{5, 5, 5}, input: 5, want: true},
		{inputSlice: []int{0, 1, 2}, input: 0, want: true},
		{inputSlice: []int{1, 2, 3}, input: 0, want: false},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, slices.ContainsInSlice(tc.inputSlice, tc.input))
	}

	casesStr := []testCase[string]{
		{inputSlice: []string{"a", "b", "c"}, input: "b", want: true},
		{inputSlice: []string{"a", "b", "c"}, input: "d", want: false},
		{inputSlice: []string{}, input: "x", want: false},
		{inputSlice: []string{"", "a"}, input: "", want: true},
		{inputSlice: []string{"a"}, input: "", want: false},
		{inputSlice: []string{"Go"}, input: "go", want: false},
	}
	for _, tc := range casesStr {
		require.Equal(t, tc.want, slices.ContainsInSlice(tc.inputSlice, tc.input))
	}

	type point struct{ X, Y int }
	p1 := point{1, 2}
	p2 := point{3, 4}
	p3 := point{5, 6}
	casesStruct := []testCase[point]{
		{inputSlice: []point{p1, p2, p3}, input: point{3, 4}, want: true},
		{inputSlice: []point{p1, p3}, input: point{3, 4}, want: false},
		{inputSlice: []point{p2, p2}, input: point{3, 4}, want: true},
	}
	for _, tc := range casesStruct {
		require.Equal(t, tc.want, slices.ContainsInSlice(tc.inputSlice, tc.input))
	}
}

func TestExcludeFromSlice(t *testing.T) {
	t.Parallel()
	src := []string{"a", "b", "c", "d", "e"}
	exclude := []string{"b", "d"}
	require.Equal(t, []string{"a", "c", "e"}, slices.ExcludeFromSlice(src, exclude))

	src = []string{"a", "", "", "c"}
	require.Equal(t, []string{"a", "c"}, slices.ExcludeFromSlice(src, []string{""}))
}

func TestChunkSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		slice     []int
		chunkSize int
		want      [][]int
	}{
		{"exact division", []int{1, 2, 3, 4, 5, 6}, 2, [][]int{{1, 2}, {3, 4}, {5, 6}}},
		{"remainder", []int{1, 2, 3, 4, 5, 6, 7, 8}, 3, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8}}},
		{"single chunk", []int{1, 2, 3}, 5, [][]int{{1, 2, 3}}},
		{"empty slice", []int{}, 3, [][]int{}},
		{"chunk size larger than slice", []int{1, 2}, 10, [][]int{{1, 2}}},
		{"zero chunk size", []int{1, 2}, 0, nil},
		{"negative chunk size", []int{1, 2}, -1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := slices.ChunkSlice(tt.slice, tt.chunkSize)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAppendBounded(t *testing.T) {
	t.Parallel()
	s := []int{1, 2, 3}
	s = slices.AppendBounded(s, 4, 5)
	require.Equal(t, []int{1, 2, 3, 4}, s)

	s = []int{1, 2, 3}
	s = slices.AppendBounded(s, 4, 3)
	require.Equal(t, []int{2, 3, 4}, s)

	s = nil
	s = slices.AppendBounded(s, 1, 2)
	s = slices.AppendBounded(s, 2, 2)
	require.Equal(t, []int{1, 2}, s)
	s = slices.AppendBounded(s, 3, 2)
	require.Equal(t, []int{2, 3}, s)

	s = []int{1}
	s = slices.AppendBounded(s, 2, 0)
	require.Equal(t, []int{1}, s)
}

func TestKeepLast(t *testing.T) {
	t.Parallel()
	require.Nil(t, slices.KeepLast([]int{1, 2, 3}, 0))
	require.Equal(t, []int{1, 2, 3}, slices.KeepLast([]int{1, 2, 3}, 10))
	require.Equal(t, []int{2, 3}, slices.KeepLast([]int{1, 2, 3}, 2))
}

func TestConcatKeepLast(t *testing.T) {
	t.Parallel()
	require.Nil(t, slices.ConcatKeepLast([][]int{{1, 2}, {3, 4}}, 0))
	require.Equal(t, []int{3, 4}, slices.ConcatKeepLast([][]int{{1, 2}, {3, 4}}, 2))
	require.Equal(t, []int{1, 2, 3, 4}, slices.ConcatKeepLast([][]int{{1, 2}, {3, 4}}, 8))
	require.Equal(t, []int{2, 3, 4, 5, 6}, slices.ConcatKeepLast([][]int{{1}, {2, 3}, {4, 5, 6}}, 5))
}
