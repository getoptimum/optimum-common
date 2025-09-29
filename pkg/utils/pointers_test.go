package utils_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPointers(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		str := uuid.NewString()
		strPointer := utils.ToPointer(str)
		require.Equal(t, str, utils.FromPointer(strPointer))
	})
	t.Run("int", func(t *testing.T) {
		integerVal := 123
		integerPointer := utils.ToPointer(integerVal)
		require.Equal(t, integerVal, utils.FromPointer(integerPointer))
	})
	t.Run("float", func(t *testing.T) {
		floatVal := 123.123
		floatPointer := utils.ToPointer(floatVal)
		require.Equal(t, floatVal, utils.FromPointer(floatPointer))
	})
	t.Run("nil", func(t *testing.T) {
		var sample *string
		require.Equal(t, "", utils.FromPointer(sample))
	})
}
