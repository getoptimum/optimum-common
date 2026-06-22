package chain_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/chain"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	table := map[string]chain.Chain{
		"1":           chain.ChainMainnet,
		" 1 ":         chain.ChainMainnet,
		"mainnet":     chain.ChainMainnet,
		"  MAINNET  ": chain.ChainMainnet,
		"hoodi":       chain.ChainHoodi,
		"HOODI":       chain.ChainHoodi,
		"560048":      chain.ChainHoodi,
	}
	for k, v := range table {
		canon, err := chain.ChainFromString(k)
		require.NoError(t, err)
		require.Equal(t, canon, v)
	}

	invalidCases := []string{
		"custom",
		"",
		"unknown",
	}
	for _, i := range invalidCases {
		v, err := chain.ChainFromString(i)
		require.Error(t, err)
		require.Equal(t, v.String(), "")
	}
}

func TestChainFromInt(t *testing.T) {
	canon, err := chain.ChainFromInt(1)
	require.NoError(t, err)
	require.Equal(t, canon, chain.ChainMainnet)
	require.Equal(t, canon.ID(), uint64(1))

	canon, err = chain.ChainFromInt(560048)
	require.NoError(t, err)
	require.Equal(t, canon, chain.ChainHoodi)
	require.Equal(t, canon.ID(), uint64(560048))

	canon, err = chain.ChainFromInt(999999)
	require.Error(t, err)
	require.Equal(t, canon.String(), "")
	require.Equal(t, canon.ID(), uint64(0))
}

func TestGenesisTime(t *testing.T) {
	table := map[string]uint64{
		"mainnet":  1606824023,
		"1":        1606824023,
		" MAINNET": 1606824023, //nolint:gocritic // it ok
		"hoodi":    1742213400,
		"HOODI":    1742213400,
		"560048":   1742213400,
	}
	for k, v := range table {
		ts, ok := chain.GenesisTime(k)
		require.True(t, ok)
		require.Equal(t, v, ts)
	}

	_, ok := chain.GenesisTime("unknown")
	require.False(t, ok)
}
