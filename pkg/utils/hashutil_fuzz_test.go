package utils_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
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
		sha256Result := utils.HashSHA256(data)
		if len(sha256Result) != 64 {
			t.Fatalf("HashSHA256: expected 64 chars, got %d", len(sha256Result))
		}

		// Test HashSHA512 - must always return 128 hex chars
		sha512Result := utils.HashSHA512(data)
		if len(sha512Result) != 128 {
			t.Fatalf("HashSHA512: expected 128 chars, got %d", len(sha512Result))
		}

		// Test HashBytes - empty returns empty, otherwise 64 hex chars
		hashBytesResult := utils.HashBytes(data)
		if len(data) == 0 {
			if hashBytesResult != "" {
				t.Fatal("HashBytes: expected empty string for empty input")
			}
		} else if len(hashBytesResult) != 64 {
			t.Fatalf("HashBytes: expected 64 chars, got %d", len(hashBytesResult))
		}

		// Test MsgHashWithTimestamp - must always return 64 hex chars
		msgHash := utils.MsgHashWithTimestamp(topic, data, timestamp)
		if len(msgHash) != 64 {
			t.Fatalf("MsgHashWithTimestamp: expected 64 chars, got %d", len(msgHash))
		}

		// Test HashSHA256String - must always return [32]byte
		rawHash := utils.HashSHA256String(data)
		if len(rawHash) != 32 {
			t.Fatalf("HashSHA256String: expected 32 bytes, got %d", len(rawHash))
		}

		// Test HashXXHash - just ensure no panic
		_ = utils.HashXXHash(data)

		// Verify determinism
		if utils.HashSHA256(data) != sha256Result {
			t.Fatal("HashSHA256 is not deterministic")
		}
	})
}
