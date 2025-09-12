package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	randMath "math/rand"
	"strconv"
)

// RandInt64 generates a random non-negative int64 in [0, math.MaxInt64]
func RandInt64() (int64, error) {
	bound := new(big.Int).Lsh(big.NewInt(1), 63) // 2^63
	n, err := rand.Int(rand.Reader, bound)
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

// RandInt generates a random positive int value
func RandInt() (int, error) {
	// Use native int width: for 32-bit => 2^31; for 64-bit => 2^63.
	bound := new(big.Int).Lsh(big.NewInt(1), uint(strconv.IntSize-1))
	n, err := rand.Int(rand.Reader, bound) // 0 <= n < 2^(bits-1)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// RandBetween returns a uniform random int in [minVal, maxVal)
func RandBetween(minVal, maxVal int) (int, error) {
	if maxVal <= minVal {
		return 0, fmt.Errorf("invalid range: minVal=%d must be < maxVal=%d", minVal, maxVal)
	}
	// span fits in native int by definition (result < maxVal), but use int64 for safe bound math
	span := int64(maxVal) - int64(minVal)             // > 0
	n, err := rand.Int(rand.Reader, big.NewInt(span)) // 0 <= n < span
	if err != nil {
		return 0, err
	}
	return minVal + int(n.Int64()), nil
}

// Shuffle shuffles a slice of any type
func Shuffle[T any](lst []T) {
	randMath.Shuffle(len(lst), func(i, j int) {
		lst[i], lst[j] = lst[j], lst[i]
	})
}

// MsgHash returns a deterministic hash of topic + message
func MsgHash(topic string, message []byte) string {
	h := sha256.New()
	h.Write([]byte(topic))
	h.Write([]byte{0}) // delimiter to avoid collisions
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}
