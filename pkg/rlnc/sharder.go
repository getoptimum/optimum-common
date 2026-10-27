package rlnc

const (
	// HeaderSize is the size of the header RLNC's sharder prepends to a payload:
	// three uint32 fields (payload, hash and padding sizes).
	HeaderSize = 12

	// PolynomialHashPlaceholderSize is the hash region RLNC's sharder reserves
	// after the payload.
	PolynomialHashPlaceholderSize = 4
)

// ShardCount returns m, the number of shards RLNC's sharder needs for a
// payload of payloadLen bytes, as ceil((payloadLen + HeaderSize +
// PolynomialHashPlaceholderSize) / maxShardSize). The sharder then rounds m up
// to a multiple of k when k > 1. It returns 0 when maxShardSize is 0.
func ShardCount(payloadLen uint64, maxShardSize uint32) uint64 {
	if maxShardSize == 0 {
		return 0
	}

	base := payloadLen + HeaderSize + PolynomialHashPlaceholderSize

	return (base + uint64(maxShardSize) - 1) / uint64(maxShardSize)
}
