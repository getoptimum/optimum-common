package chain_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/chain"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	require.Equal(t, "mainnet", chain.Normalize("1"))
	require.Equal(t, "mainnet", chain.Normalize("mainnet"))
	require.Equal(t, "mainnet", chain.Normalize("  MAINNET  "))
	require.Equal(t, "hoodi", chain.Normalize("hoodi"))
	require.Equal(t, "hoodi", chain.Normalize("HOODI"))
	require.Equal(t, "hoodi", chain.Normalize("560048"))
	require.Equal(t, "custom", chain.Normalize("custom"))
}

func TestNormalizeChainID_withCustomMappings(t *testing.T) {
	mappings := map[string]string{
		"1":       "mainnet",
		"mainnet": "mainnet",
		"custom":  "custom_chain",
	}
	require.Equal(t, "mainnet", chain.NormalizeChainID("1", mappings))
	require.Equal(t, "custom_chain", chain.NormalizeChainID("custom", mappings))
	require.Equal(t, "unknown", chain.NormalizeChainID("unknown", mappings))
}

func TestNormalizeChainID_nilUsesDefaults(t *testing.T) {
	require.Equal(t, "mainnet", chain.NormalizeChainID("1", nil))
}
