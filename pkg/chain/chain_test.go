package chain_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/chain"
	"github.com/stretchr/testify/require"
)

func TestChainFromString(t *testing.T) {
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
		require.Equal(t, v, canon)
	}

	invalidCases := []string{
		"custom",
		"",
		"unknown",
	}
	for _, i := range invalidCases {
		v, err := chain.ChainFromString(i)
		require.Error(t, err)
		require.Equal(t, "", v.String())
	}
}

func TestChainFromInt(t *testing.T) {
	canon, err := chain.ChainFromInt(1)
	require.NoError(t, err)
	require.Equal(t, chain.ChainMainnet, canon)
	require.Equal(t, uint64(1), canon.ID())

	canon, err = chain.ChainFromInt(560048)
	require.NoError(t, err)
	require.Equal(t, chain.ChainHoodi, canon)
	require.Equal(t, uint64(560048), canon.ID())

	canon, err = chain.ChainFromInt(999999)
	require.Error(t, err)
	require.Equal(t, "", canon.String())
	require.Equal(t, uint64(0), canon.ID())
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

func TestParseChainIDParam(t *testing.T) {
	cases := []struct {
		raw       string
		allowZero bool
		want      uint64
		wantErr   bool
	}{
		{"", false, 0, false},
		{"   ", false, 0, false},
		{"1", false, 1, false},
		{" 1 ", false, 1, false},
		{"18446744073709551615", false, 18446744073709551615, false},
		{"0", false, 0, true},
		{"00", false, 0, true},
		{"-1", false, 0, true},
		{"+1", false, 0, true},
		{"abc", false, 0, true},
		{"1.0", false, 0, true},
		{"18446744073709551616", false, 0, true},
		{"", true, 0, false},
		{"0", true, 0, false},
		{"00", true, 0, false},
		{"1", true, 1, false},
		{"-1", true, 0, true},
		{"abc", true, 0, true},
		{"18446744073709551616", true, 0, true},
	}
	for _, c := range cases {
		id, err := chain.ParseChainIDParam(c.raw, c.allowZero)
		if c.wantErr {
			require.Error(t, err)
			require.Equal(t, uint64(0), id)
			continue
		}
		require.NoError(t, err)
		require.Equal(t, c.want, id)
	}
}
