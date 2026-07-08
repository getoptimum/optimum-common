package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math"
	"sync"

	"github.com/cespare/xxhash/v2"
)

var (
	sha256Pool = sync.Pool{New: func() any {
		return sha256.New()
	}}

	sha512Pool = sync.Pool{New: func() any {
		return sha512.New()
	}}
)

func calcSHA256hash(data []byte) []byte {
	h, ok := sha256Pool.Get().(hash.Hash)
	if !ok {
		h = sha256.New()
	}
	defer sha256Pool.Put(h)
	h.Reset()

	// The hash interface never returns an error, for that reason
	// we are not handling the error below. For reference, it is
	// stated here https://golang.org/pkg/hash/#Hash

	// #nosec G104
	_, _ = h.Write(data)
	var sum [sha256.Size]byte
	h.Sum(sum[:0])
	return sum[:]
}

func calcSHA512hash(data []byte) []byte {
	h, ok := sha512Pool.Get().(hash.Hash)
	if !ok {
		h = sha512.New()
	}
	defer sha512Pool.Put(h)
	h.Reset()

	// The hash interface never returns an error, for that reason
	// we are not handling the error below. For reference, it is
	// stated here https://golang.org/pkg/hash/#Hash

	// #nosec G104
	_, _ = h.Write(data)
	var sum [sha512.Size]byte
	h.Sum(sum[:0])
	return sum[:]
}

// SHA256 computes SHA-256 hash returns hex-encoded string
func SHA256(data []byte) string {
	return hex.EncodeToString(calcSHA256hash(data))
}

// SHA512 computes SHA-512 hash, returns hex-encoded string
func SHA512(data []byte) string {
	return hex.EncodeToString(calcSHA512hash(data))
}

// SHA256String computes the SHA-256 hash, returns the raw 32-byte array
func SHA256String(data []byte) [sha256.Size]byte {
	return [32]byte(calcSHA256hash(data))
}

// XXHash computes the XXHash, fast, non-cryptographic
func XXHash(data []byte) uint64 {
	return xxhash.Sum64(data)
}

// BytesHash returns a SHA256 hash of the given bytes as a hex string.
// Returns empty string if input is empty.
func BytesHash(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return SHA256(b)
}

// WriteBool writes a boolean value to the hash in a deterministic format (1 for true, 0 for false).
func WriteBool(h hash.Hash, v bool) {
	if v {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
}

// WriteInt64 writes an int64 value to the hash in little-endian format.
func WriteInt64(h hash.Hash, v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v)) //nolint:gosec // uint64 conversion is safe
	h.Write(buf[:])
}

// WriteFloat32 writes a float32 value to the hash by converting it to its IEEE 754 binary representation.
func WriteFloat32(h hash.Hash, v float32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
	h.Write(buf[:])
}
