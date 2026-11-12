package chain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

type Chain string

const (
	ChainMainnet Chain = "mainnet"
	ChainHoodi   Chain = "hoodi"
	ChainSepolia Chain = "sepolia"
)

func (c Chain) String() string {
	return string(c)
}

// ID returns the numeric chain id, or 0 when the chain is not registered.
func (c Chain) ID() uint64 {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if id, ok := defaultChainMappingID[c]; ok {
		return id
	}
	return 0
}

var (
	// registryMu guards the maps below: built-in entries plus chains added
	// via RegisterChain at service startup.
	registryMu sync.RWMutex

	defaultChainMapping = map[string]Chain{
		"1":        ChainMainnet,
		"mainnet":  ChainMainnet,
		"560048":   ChainHoodi,
		"hoodi":    ChainHoodi,
		"11155111": ChainSepolia,
		"sepolia":  ChainSepolia,
	}
	defaultChainIDMapping = map[uint64]Chain{
		1:        ChainMainnet,
		560048:   ChainHoodi,
		11155111: ChainSepolia,
	}
	defaultChainMappingID = map[Chain]uint64{
		ChainMainnet: 1,
		ChainHoodi:   560048,
		ChainSepolia: 11155111,
	}
	// Beacon chain genesis timestamps (unix seconds) per canonical chain ID.
	genesisTimestamps = map[Chain]uint64{
		ChainMainnet: 1606824023, // Dec 1, 2020 12:00:23 UTC
		ChainHoodi:   1742213400, // Mar 17, 2025 12:10:00 UTC
		ChainSepolia: 1655733600, // Jun 20, 2022 14:00:00 UTC
	}
)

// RegisterChain adds a chain supplied via service config (a devnet such as
// Kurtosis, or a public testnet like Platåberget) so the package-level lookups
// resolve it. Registering an identical triple again is a no-op; a name or id
// that is already mapped to different values is rejected. Safe for concurrent use.
func RegisterChain(name string, chainID, genesisTime uint64) (Chain, error) {
	s := strings.TrimSpace(strings.ToLower(name))
	switch {
	case s == "":
		return "", errors.New("empty chain name")
	case chainID == 0:
		return "", errors.New("chain ID must be non-zero")
	case genesisTime == 0:
		return "", errors.New("genesis time must be non-zero")
	case genesisTime > math.MaxInt64:
		// Consumers convert genesis to time.Unix(int64, 0).
		return "", errors.New("genesis time exceeds int64 range")
	}
	// Numeric names would be ambiguous with chain-id string lookups.
	if _, err := strconv.ParseUint(s, 10, 64); err == nil {
		return "", fmt.Errorf("chain name %q must not be numeric", name)
	}

	c := Chain(s)

	registryMu.Lock()
	defer registryMu.Unlock()

	if existingID, ok := defaultChainMappingID[c]; ok {
		if existingID == chainID && genesisTimestamps[c] == genesisTime {
			return c, nil
		}
		return "", fmt.Errorf("chain %q is already registered with different parameters", s)
	}
	if existing, ok := defaultChainIDMapping[chainID]; ok {
		return "", fmt.Errorf("chain ID %d is already registered as %q", chainID, existing)
	}

	defaultChainMapping[s] = c
	defaultChainMapping[strconv.FormatUint(chainID, 10)] = c
	defaultChainIDMapping[chainID] = c
	defaultChainMappingID[c] = chainID
	genesisTimestamps[c] = genesisTime
	return c, nil
}

// GenesisTime returns the beacon chain genesis unix timestamp for the given
// chain. The input is normalized, so "hoodi", "HOODI", and "560048" all work.
func GenesisTime(chainID string) (uint64, bool) {
	canon, err := ChainFromString(chainID)
	if err != nil {
		return 0, false
	}
	registryMu.RLock()
	defer registryMu.RUnlock()
	t, ok := genesisTimestamps[canon]
	return t, ok
}

func ChainFromString(chainID string) (Chain, error) {
	s := strings.TrimSpace(strings.ToLower(chainID))
	if s == "" {
		return "", errors.New("empty chain ID")
	}

	registryMu.RLock()
	defer registryMu.RUnlock()
	if canon, ok := defaultChainMapping[s]; ok {
		return canon, nil
	}
	return "", fmt.Errorf("unknown chain ID: %q", chainID)
}

func ChainFromInt(chainID uint64) (Chain, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if canon, ok := defaultChainIDMapping[chainID]; ok {
		return canon, nil
	}
	return "", fmt.Errorf("unknown chain ID: %d", chainID)
}

// ParseChainIDParam parses a chain_id query param. Empty means 0 ("all chains").
// A literal 0 is rejected unless allowZero is true.
func ParseChainIDParam(raw string, allowZero bool) (uint64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, nil
	}
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid chain_id %q: %w", raw, err)
	}
	if id == 0 && !allowZero {
		return 0, errors.New("chain_id must be a non-zero uint64")
	}
	return id, nil
}
