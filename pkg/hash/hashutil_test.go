package hash_test

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/hash"
	"github.com/stretchr/testify/require"
)

func TestHashSHA512(t *testing.T) {
	table := map[string]string{
		"Hello, World!": "374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6cc69291e0fa2fe0006a52570ef18c19def4e617c33ce52ef0a6e5fbe318cb0387",
		"Hello, 123!":   "c242f458c7473bb510c3d273e593f4d16ea4d45a1b763308ae4e780c5e6175afdfc3b2735ea81a9a9b84a8b5515749d5d443f641d7aed13236295303ecf420b5",
		"Hello, 123":    "84df6bdafdaa325beeaa4dedf46e6519e350ac7c9936d44d5b1de84359572d3d7047bece9f25dbb12876b9f307bb994f0df737b87757a0081583f3b23b7d4a4b",
	}
	for src, res := range table {
		require.Equal(t, res, hash.HashSHA512([]byte(src)))
	}
}

func TestHashSHA256(t *testing.T) {
	table := map[string]string{
		"Hello, World!": "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		"Hello, 123!":   "52ce7f8d6d1f955e01b896c8fb38de421b4f9d0a2978fb1e3a3c9f3a6efa80ff",
		"Hello, 123":    "30b6bfae65bce9ae9ab1cef925407ddc3bcc3ee3ccbb4991619a4d7cd0c72675",
	}
	for src, res := range table {
		require.Equal(t, res, hash.HashSHA256([]byte(src)))
	}
}

func BenchmarkHashXXHash(b *testing.B) {
	data := []byte("Hello, World!")
	for i := 0; i < b.N; i++ {
		hash.HashXXHash(data)
	}
}

func BenchmarkHashSHA256String(b *testing.B) {
	data := []byte("Hello, World!")
	for i := 0; i < b.N; i++ {
		hash.HashSHA256String(data)
	}
}

func TestHashBytes(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "simple string",
			input:    []byte("Hello, World!"),
			expected: "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0x01, 0x02, 0x03},
			expected: "054edec1d0211f624fed0cbca9d4f9400b0e491c43742af2c5b0abebf0c990d8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := hash.HashBytes(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestHashBytes_MatchesHashSHA256(t *testing.T) {
	data := []byte("test data")
	hashBytes := hash.HashBytes(data)
	hashSHA256 := hash.HashSHA256(data)
	require.Equal(t, hashSHA256, hashBytes, "HashBytes should match HashSHA256 for non-empty input")
}

func TestPooledHashes_ConcurrentCorrectness(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		payload []byte
	}{
		{name: "empty", payload: []byte{}},
		{name: "text", payload: []byte("Hello, World!")},
		{name: "binary", payload: []byte{0x00, 0x01, 0x02, 0x03, 0xff}},
	}

	largePayload, err := json.Marshal(struct {
		Topic string `json:"topic"`
		Body  string `json:"body"`
		Count int    `json:"count"`
	}{
		Topic: "large-payload",
		Body:  string(make([]byte, 2048)),
		Count: 2048,
	})
	require.NoError(t, err)
	testCases = append(testCases, struct {
		name    string
		payload []byte
	}{
		name:    "large",
		payload: largePayload,
	})

	const goroutines = 16
	const iterations = 100

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expected256Array := sha256.Sum256(tc.payload)
			expected512Array := sha512.Sum512(tc.payload)
			expected256Hex := hex.EncodeToString(expected256Array[:])
			expected512Hex := hex.EncodeToString(expected512Array[:])

			var wg sync.WaitGroup
			errCh := make(chan error, goroutines*iterations*3)
			for range goroutines {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for range iterations {
						if got := hash.HashSHA256(tc.payload); got != expected256Hex {
							errCh <- fmt.Errorf("HashSHA256 mismatch: got %q want %q", got, expected256Hex)
						}
						if got := hash.HashSHA512(tc.payload); got != expected512Hex {
							errCh <- fmt.Errorf("HashSHA512 mismatch: got %q want %q", got, expected512Hex)
						}
						if got := hash.HashSHA256String(tc.payload); got != expected256Array {
							errCh <- fmt.Errorf("HashSHA256String mismatch: got %x want %x", got, expected256Array)
						}
					}
				}()
			}
			wg.Wait()
			close(errCh)

			for err := range errCh {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgHashWithTimestamp_Determinism(t *testing.T) {
	t.Parallel()

	topic := "topic"
	msg := []byte("hello")
	timestamp := int64(1234567890)

	h1 := hash.MsgHashWithTimestamp(topic, msg, timestamp)
	h2 := hash.MsgHashWithTimestamp(topic, msg, timestamp)
	require.Equal(t, h1, h2, "same inputs should yield same hash")

	// Different timestamp should produce different hash
	h3 := hash.MsgHashWithTimestamp(topic, msg, timestamp+1)
	require.NotEqual(t, h1, h3, "different timestamp should produce different hash")

	// Same message with different timestamp should be different
	h4 := hash.MsgHashWithTimestamp(topic, msg, timestamp-1)
	require.NotEqual(t, h1, h4, "different timestamp should produce different hash")
}

func TestMsgHashWithTimestamp_DifferentFromMsgHash(t *testing.T) {
	t.Parallel()

	topic := "topic"
	msg := []byte("hello")
	timestamp := int64(1234567890)

	hashWithTS := hash.MsgHashWithTimestamp(topic, msg, timestamp)
	hashWithoutTS := hash.MsgHash(topic, msg)

	require.NotEqual(t, hashWithTS, hashWithoutTS, "timestamp should affect hash")
}

func TestMsgHashWithTimestamp_IsHex_And_Length(t *testing.T) {
	t.Parallel()

	h := hash.MsgHashWithTimestamp("t", []byte("m"), 1234567890)
	require.NotEmpty(t, h)

	// expect hex
	_, err := hex.DecodeString(h)
	require.NoError(t, err, "hash must be hex-encoded")
	require.Len(t, h, 64, "unexpected hash length (expecting SHA-256 hex)")
}
