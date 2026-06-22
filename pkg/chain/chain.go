package chain

import (
	"errors"
	"fmt"
	"strings"
)

type Chain string

const (
	ChainMainnet Chain = "mainnet"
	ChainHoodi   Chain = "hoodi"
)

func (c Chain) String() string {
	return string(c)
}

func (c Chain) ID() uint64 {
	if id, ok := defaultChainMappingID[c]; ok {
		return id
	}
	return 0
}

var (
	defaultChainMapping = map[string]Chain{
		"1":       ChainMainnet,
		"mainnet": ChainMainnet,
		"560048":  ChainHoodi,
		"hoodi":   ChainHoodi,
	}
	defaultChainIDMapping = map[uint64]Chain{
		1:      ChainMainnet,
		560048: ChainHoodi,
	}
	defaultChainMappingID = map[Chain]uint64{
		ChainMainnet: 1,
		ChainHoodi:   560048,
	}
	// Beacon chain genesis timestamps (unix seconds) per canonical chain ID.
	genesisTimestamps = map[Chain]uint64{
		ChainMainnet: 1606824023, // Dec 1, 2020 12:00:23 UTC
		ChainHoodi:   1742213400, // Mar 17, 2025 12:10:00 UTC
	}
)

// GenesisTime returns the beacon chain genesis unix timestamp for the given
// chain. The input is normalized, so "hoodi", "HOODI", and "560048" all work.
func GenesisTime(chainID string) (uint64, bool) {
	canon, err := ChainFromString(chainID)
	if err != nil {
		return 0, false
	}
	t, ok := genesisTimestamps[canon]
	return t, ok
}

func ChainFromString(chainID string) (Chain, error) {
	s := strings.TrimSpace(strings.ToLower(chainID))
	if s == "" {
		return "", errors.New("empty chain ID")
	}

	if canon, ok := defaultChainMapping[s]; ok {
		return canon, nil
	}
	return "", errors.New("unknown chain ID: " + chainID)
}

func ChainFromInt(chainID uint64) (Chain, error) {
	if canon, ok := defaultChainIDMapping[chainID]; ok {
		return canon, nil
	}
	return "", fmt.Errorf("unknown chain ID: %d", chainID)
}
