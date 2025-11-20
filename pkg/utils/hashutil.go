package utils

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math"

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

// HashBytes returns a SHA256 hash of the given bytes as a hex string.
// Returns empty string if input is empty.
func HashBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func WriteBool(h hash.Hash, v bool) {
	if v {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
}
func WriteInt64(h hash.Hash, v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v)) //nolint:gosec // uint64 conversion is safe
	h.Write(buf[:])
}
func WriteFloat32(h hash.Hash, v float32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
	h.Write(buf[:])
}
