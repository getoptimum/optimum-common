package rlnc

import (
	"bytes"
	"crypto/subtle"
)

// GFMultiply multiplies two elements of the RLNC encoder's GF(256) field, which uses
// the primitive polynomial 0x11d.
func GFMultiply(left, right byte) byte {
	var product byte
	for right != 0 {
		if right&1 != 0 {
			product ^= left
		}

		carry := left&0x80 != 0
		left <<= 1

		if carry {
			left ^= 0x1d
		}

		right >>= 1
	}

	return product
}

// GFInverse returns the multiplicative inverse of value in GF(256). Zero has no
// inverse and returns zero.
func GFInverse(value byte) byte {
	if value == 0 {
		return 0
	}

	// a^254 is a^-1 in GF(256).
	result := byte(1)
	base := value
	for exponent := 254; exponent > 0; exponent >>= 1 {
		if exponent&1 != 0 {
			result = GFMultiply(result, base)
		}

		base = GFMultiply(base, base)
	}

	return result
}

// Scale multiplies every element of row by factor in place.
func Scale(row []byte, factor byte) {
	for index := range row {
		row[index] = GFMultiply(factor, row[index])
	}
}

// XorScaled adds factor*src to dst in place. Addition in GF(256) is XOR. src
// must be at least as long as dst.
func XorScaled(dst, src []byte, factor byte) {
	for index := range dst {
		dst[index] ^= GFMultiply(factor, src[index])
	}
}

// LeadingColumn returns the index of row's first nonzero element, or -1 if row
// is all zeros.
func LeadingColumn(row []byte) int {
	for index, value := range row {
		if value != 0 {
			return index
		}
	}

	return -1
}

// auditBlockSize is how many data bytes AuditSymbol checks at a time. A block
// fits on the stack, so the audit allocates nothing, and covers a 4 KiB shard
// in one pass.
const auditBlockSize = 4096

// mulTable holds every GF(256) product, so AuditSymbol costs one lookup per
// coefficient and data byte instead of a bitwise multiply.
var mulTable = func() (table [256][256]byte) {
	for left := range 256 {
		for right := range 256 {
			table[left][right] = GFMultiply(byte(left), byte(right))
		}
	}

	return table
}()

// AuditSymbol reports whether data is the combination of shards that
// coefficients describes. shards are the k verified source shards of one
// generation. A symbol with other than k coefficients, or with data of another
// length than the shards, fails the audit. It does not modify its arguments.
func AuditSymbol(coefficients, data []byte, shards [][]byte) (valid bool) {
	if len(data) == 0 || len(shards) == 0 || len(coefficients) != len(shards) {
		return false
	}

	for _, shard := range shards {
		if len(shard) != len(data) {
			return false
		}
	}

	// Accumulating one shard at a time reads every shard sequentially. Zero
	// coefficients add nothing, and a coefficient of 1 is a plain XOR, so a
	// source symbol costs one XOR pass.
	var block [auditBlockSize]byte
	for start := 0; start < len(data); start += auditBlockSize {
		end := min(start+auditBlockSize, len(data))
		acc := block[:end-start]
		clear(acc)

		for row, coefficient := range coefficients {
			src := shards[row][start:end]

			switch coefficient {
			case 0:
			case 1:
				subtle.XORBytes(acc, acc, src)
			default:
				products := &mulTable[coefficient]
				for index, value := range src {
					acc[index] ^= products[value]
				}
			}
		}

		if !bytes.Equal(acc, data[start:end]) {
			return false
		}
	}

	return true
}
