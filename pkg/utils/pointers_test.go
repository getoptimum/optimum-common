package utils_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestToAndFromPointer(t *testing.T) {
	type testCase[T comparable] struct {
		name     string
		input    T
		isNil    bool
		expected T
	}

	t.Run("string", func(t *testing.T) {
		str := uuid.NewString()
		tests := []testCase[string]{
			{"non-nil string", str, false, str},
			{"nil string", "", true, ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var ptr *string
				if !tt.isNil {
					ptr = utils.ToPointer(tt.input)
				}
				result := utils.FromPointer(ptr)
				require.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("int", func(t *testing.T) {
		tests := []testCase[int]{
			{"non-nil int", 123, false, 123},
			{"nil int", 0, true, 0},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var ptr *int
				if !tt.isNil {
					ptr = utils.ToPointer(tt.input)
				}
				result := utils.FromPointer(ptr)
				require.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("float", func(t *testing.T) {
		tests := []testCase[float64]{
			{"non-nil float", 123.123, false, 123.123},
			{"nil float", 0.0, true, 0.0},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var ptr *float64
				if !tt.isNil {
					ptr = utils.ToPointer(tt.input)
				}
				result := utils.FromPointer(ptr)
				require.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("bool", func(t *testing.T) {
		tests := []testCase[bool]{
			{"true bool", true, false, true},
			{"false bool", false, false, false},
			{"nil bool", false, true, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var ptr *bool
				if !tt.isNil {
					ptr = utils.ToPointer(tt.input)
				}
				result := utils.FromPointer(ptr)
				require.Equal(t, tt.expected, result)
			})
		}
	})
}
