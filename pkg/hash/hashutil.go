package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	stdhash "hash"
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

// MsgHash returns a deterministic hash of topic + message for message identity (e.g. deduplication).
// For time-scoped identity (e.g. same message at different times), use MsgHashWithTimestamp.
func MsgHash(topic string, message []byte) string {
	h := sha256.New()
	h.Write([]byte(topic))
	h.Write([]byte{0}) // delimiter to avoid collisions
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

// MsgHashWithTimestamp returns a hash of topic + message + timestamp for time-scoped identity
// (e.g. duplicate detection when the same content can be sent at different times).
// For content-only identity, use MsgHash.
func MsgHashWithTimestamp(topic string, message []byte, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(topic))
	h.Write(message)
	// Convert timestamp to bytes and include in hash
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp)) //nolint:gosec // int64 to uint64 conversion is safe for timestamps
	h.Write(timestampBytes)
	return hex.EncodeToString(h.Sum(nil))
}

// WriteBool writes a boolean value to the hash in a deterministic format (1 for true, 0 for false).
func WriteBool(h stdhash.Hash, v bool) {
	if v {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
}

// WriteInt64 writes an int64 value to the hash in little-endian format.
func WriteInt64(h stdhash.Hash, v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v)) //nolint:gosec // uint64 conversion is safe
	h.Write(buf[:])
}

// WriteFloat32 writes a float32 value to the hash by converting it to its IEEE 754 binary representation.
func WriteFloat32(h stdhash.Hash, v float32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
	h.Write(buf[:])
}
