package rlnc

// SharderMetadataSize is the header and reserved polynomial-hash region added
// by RLNC's sharder module before it calculates its shard count. Keep this
// synchronized with that package's HeaderSize (12) and PolynomialHashPlaceholderSize (4).
const SharderMetadataSize = 16
