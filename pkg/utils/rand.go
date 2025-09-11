package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	randMath "math/rand"
	"time"
)

// RandInt64 generates a random int64 value
// uses crypto/rand to generate random number
// ensures the generated number is within the range of int64
func RandInt64() int64 {
	for {
		dat := make([]byte, 8)
		_, _ = rand.Read(dat)
		intValue := binary.BigEndian.Uint64(dat)
		if intValue > math.MaxInt64 {
			continue
		}
		return int64(intValue)
	}
}

// RandInt generates a random int value
func RandInt() int {
	for {
		dat := make([]byte, 4)
		_, _ = rand.Read(dat)
		intValue := binary.BigEndian.Uint32(dat)
		if intValue > math.MaxInt32 {
			continue
		}
		return int(intValue)
	}
}

// RandBetween generates random int value between minVal and maxVal
func RandBetween(minVal, maxVal int) int {
	r := RandInt()
	return r%(maxVal-minVal) + minVal
}

// Shuffle shuffles a slice of any type
func Shuffle[T any](lst []T) {
	//nolint:gosec
	randMath.New(randMath.NewSource(time.Now().UnixNano()))
	randMath.Shuffle(len(lst), func(i, j int) {
		lst[i], lst[j] = lst[j], lst[i]
	})
}

// MsgHash returns a deterministic hash of topic + message
func MsgHash(topic string, message []byte) string {
	h := sha256.New()
	h.Write([]byte(topic))
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}
