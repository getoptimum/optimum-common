package chain_test

import (
	"fmt"
	"math"
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
		"sepolia":     chain.ChainSepolia,
		"SEPOLIA":     chain.ChainSepolia,
		"11155111":    chain.ChainSepolia,
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

	canon, err = chain.ChainFromInt(11155111)
	require.NoError(t, err)
	require.Equal(t, chain.ChainSepolia, canon)
	require.Equal(t, uint64(11155111), canon.ID())

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
		"sepolia":  1655733600,
		"11155111": 1655733600,
	}
	for k, v := range table {
		ts, ok := chain.GenesisTime(k)
		require.True(t, ok)
		require.Equal(t, v, ts)
	}

	_, ok := chain.GenesisTime("unknown")
	require.False(t, ok)
}

func TestRegisterChain(t *testing.T) {
	// ethereum-package's default network id, a realistic devnet shape.
	canon, err := chain.RegisterChain(" Kurtosis-Dev ", 3151908, 1700000000)
	require.NoError(t, err)
	require.Equal(t, chain.Chain("kurtosis-dev"), canon)
	require.Equal(t, uint64(3151908), canon.ID())

	byName, err := chain.ChainFromString("KURTOSIS-DEV")
	require.NoError(t, err)
	require.Equal(t, canon, byName)

	byIDString, err := chain.ChainFromString("3151908")
	require.NoError(t, err)
	require.Equal(t, canon, byIDString)

	byInt, err := chain.ChainFromInt(3151908)
	require.NoError(t, err)
	require.Equal(t, canon, byInt)

	ts, ok := chain.GenesisTime("kurtosis-dev")
	require.True(t, ok)
	require.Equal(t, uint64(1700000000), ts)
	ts, ok = chain.GenesisTime("3151908")
	require.True(t, ok)
	require.Equal(t, uint64(1700000000), ts)

	// Re-registering the identical triple is a no-op.
	again, err := chain.RegisterChain("kurtosis-dev", 3151908, 1700000000)
	require.NoError(t, err)
	require.Equal(t, canon, again)

	// Conflicts are rejected.
	_, err = chain.RegisterChain("kurtosis-dev", 3151909, 1700000000)
	require.Error(t, err)
	_, err = chain.RegisterChain("kurtosis-dev", 3151908, 1700000001)
	require.Error(t, err)
	_, err = chain.RegisterChain("other-name", 3151908, 1700000000)
	require.Error(t, err)
	_, err = chain.RegisterChain("mainnet", 2, 1606824023)
	require.Error(t, err)

	// Built-ins re-registered with identical values are a no-op.
	same, err := chain.RegisterChain("mainnet", 1, 1606824023)
	require.NoError(t, err)
	require.Equal(t, chain.ChainMainnet, same)

	// Invalid inputs.
	for _, c := range []struct {
		name    string
		id      uint64
		genesis uint64
	}{
		{"", 42, 1700000000},
		{"   ", 42, 1700000000},
		{"123456", 42, 1700000000},
		{"devnet", 0, 1700000000},
		{"devnet", 42, 0},
		{"devnet", 42, math.MaxUint64},
	} {
		_, err := chain.RegisterChain(c.name, c.id, c.genesis)
		require.Error(t, err, "name=%q id=%d genesis=%d", c.name, c.id, c.genesis)
	}

	// Lookups stay safe while registrations write to the maps (-race enforced).
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 200 {
			_, _ = chain.ChainFromString("mainnet")
			_, _ = chain.GenesisTime("sepolia")
			_, _ = chain.ChainFromInt(1)
		}
	}()
	for i := range 200 {
		_, err := chain.RegisterChain(fmt.Sprintf("race-%d", i), uint64(9_000_000+i), 1700000000)
		require.NoError(t, err)
	}
	<-done
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
