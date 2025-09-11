package hashutil

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"github.com/cespare/xxhash/v2"
)

// HashSHA256 computes SHA-256 hash returns hex-encoded string
func HashSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// HashSHA512 computes SHA-512 hash, returns hex-encoded string
func HashSHA512(data []byte) string {
	sum := sha512.Sum512(data)
	return hex.EncodeToString(sum[:])
}

// HashSHA256String computes the SHA-256 hash, returns the raw 32-byte array
func HashSHA256String(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// HashXXHash computes the XXHash, fast, non-cryptographic
func HashXXHash(data []byte) uint64 {
	return xxhash.Sum64(data)
}
