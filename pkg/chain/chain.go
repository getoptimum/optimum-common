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
