package rlnc

import (
	"bytes"
	"crypto/subtle"
)

// auditBlockSize is how many data bytes AuditSymbol checks at a time. A block
// fits on the stack, so the audit allocates nothing, and covers a 4 KiB shard
// in one pass.
const auditBlockSize = 4096

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
