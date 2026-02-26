package chain

import "strings"

// DefaultChainMappings returns input->canonical chain ID mappings.
func DefaultChainMappings() map[string]string {
	return map[string]string{
		"1":       "mainnet",
		"mainnet": "mainnet",
		"560048":  "hoodi",
		"hoodi":   "hoodi",
	}
}

// Beacon chain genesis timestamps (unix seconds) per canonical chain ID.
var genesisTimestamps = map[string]uint64{
	"mainnet": 1606824023, // Dec 1, 2020 12:00:23 UTC
	"hoodi":   1742213400, // Mar 17, 2025 12:10:00 UTC
}

// GenesisTime returns the beacon chain genesis unix timestamp for the given
// chain. The input is normalized, so "hoodi", "HOODI", and "560048" all work.
func GenesisTime(chainID string) (uint64, bool) {
	t, ok := genesisTimestamps[Normalize(chainID)]
	return t, ok
}

// NormalizeChainID normalizes chainID using mappings. Uses DefaultChainMappings if mappings is nil.
func NormalizeChainID(chainID string, mappings map[string]string) string {
	s := strings.TrimSpace(strings.ToLower(chainID))
	if s == "" {
		return s
	}
	if mappings == nil {
		mappings = DefaultChainMappings()
	}
	if canon, ok := mappings[s]; ok {
		return canon
	}
	return s
}

// Normalize normalizes chainID using DefaultChainMappings.
func Normalize(chainID string) string {
	return NormalizeChainID(chainID, nil)
}
