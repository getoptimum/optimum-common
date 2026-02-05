package maps_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/maps"
	"github.com/stretchr/testify/require"
)

func TestMaps(t *testing.T) {
	m := map[int]string{228: "t", 1489: "e", 666: "s", 777: "t"}
	keys := maps.MapKeys(m)
	values := maps.MapValues(m)
	require.Equal(t, len(m), len(keys))
	require.Equal(t, len(m), len(values))
	sort.Ints(keys)
	sort.Strings(values)
	require.Equal(t, []int{228, 666, 777, 1489}, keys)
	require.Equal(t, "estt", strings.Join(values, ""))
}
