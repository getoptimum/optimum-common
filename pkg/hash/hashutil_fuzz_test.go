package hash_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/hash"
	"github.com/stretchr/testify/require"
)

// FuzzHashing tests all hash functions with arbitrary input
func FuzzHashing(f *testing.F) {
	f.Add([]byte{}, "topic", int64(0))
	f.Add([]byte{0x00}, "test", int64(1000000000))
	f.Add([]byte("hello world"), "/eth2/attestation", int64(1700000000))
	f.Add(make([]byte, 1024), "", int64(-1))
	f.Add([]byte(`{"slot":123}`), "topic", int64(9223372036854775807))

	f.Fuzz(func(t *testing.T, data []byte, topic string, timestamp int64) {
		// Test HashSHA256 - must always return 64 hex chars
		sha256Result := hash.SHA256(data)
		require.Len(t, sha256Result, 64, "HashSHA256: expected 64 chars")

		// Test HashSHA512 - must always return 128 hex chars
		sha512Result := hash.SHA512(data)
		require.Len(t, sha512Result, 128, "HashSHA512: expected 128 chars")

		// Test HashBytes - empty returns empty, otherwise matches SHA-256 hex output
		hashBytesResult := hash.BytesHash(data)
		if len(data) == 0 {
			require.Empty(t, hashBytesResult, "HashBytes: expected empty string for empty input")
		} else {
			require.Equal(t, sha256Result, hashBytesResult, "HashBytes: expected HashSHA256 output")
		}

		// Test MsgHashWithTimestamp - must always return 64 hex chars
		msgHash := hash.MsgHashWithTimestamp(topic, data, timestamp)
		require.Len(t, msgHash, 64, "MsgHashWithTimestamp: expected 64 chars")

		// Test HashSHA256String - must always return [32]byte
		rawHash := hash.SHA256String(data)
		require.Len(t, rawHash, 32, "HashSHA256String: expected 32 bytes")

		// Test HashXXHash - just ensure no panic
		_ = hash.XXHash(data)

		// Verify determinism
		require.Equal(t, sha256Result, hash.SHA256(data), "HashSHA256 is not deterministic")
	})
}
