package utils_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestMaps(t *testing.T) {
	m := map[int]string{228: "t", 1488: "e", 666: "s", 777: "t"}
	keys := utils.Keys(m)
	values := utils.Values(m)
	require.Equal(t, len(keys), len(m))
	require.Equal(t, len(values), len(m))
	sort.Ints(keys)
	sort.Strings(values)
	require.Equal(t, keys, []int{228, 666, 777, 1488})
	require.Equal(t, strings.Join(values, ""), "estt")
}
