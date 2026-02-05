package hash_test

import (
	"bytes"
	"encoding/hex"
	"regexp"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/hash"
	"github.com/stretchr/testify/require"
)

func TestMsgHash_DeterminismAndOrder(t *testing.T) {
	t.Parallel()

	topic := "topic"
	msg := []byte("hello")

	h1 := hash.MsgHash(topic, msg)
	h2 := hash.MsgHash(topic, msg)
	require.Equal(t, h1, h2, "same inputs should yield same hash")

	// ensure order
	h3 := hash.MsgHash("topic2", msg)
	require.NotEqual(t, h1, h3)

	h4 := hash.MsgHash(topic, []byte("hello!"))
	require.NotEqual(t, h1, h4)
}

func TestMsgHash_DelimiterAvoidsAmbiguity(t *testing.T) {
	t.Parallel()

	aTopic, aMsg := "ab", []byte("c")
	bTopic, bMsg := "a", []byte("bc")

	ha := hash.MsgHash(aTopic, aMsg)
	hb := hash.MsgHash(bTopic, bMsg)
	require.NotEqualf(t, ha, hb, "delimiter-safe hashing should differ, but both were %s", ha)
}

func TestMsgHash_DelimiterAvoidsOtherAmbiguities(t *testing.T) {
	t.Parallel()

	// "abc" + "" vs "ab" + "c"
	ha := hash.MsgHash("abc", []byte(""))
	hb := hash.MsgHash("ab", []byte("c"))
	require.NotEqual(t, ha, hb)
}

func TestMsgHash_IsHex_And_Length(t *testing.T) {
	t.Parallel()

	h := hash.MsgHash("t", []byte("m"))
	require.NotEmpty(t, h)

	// expect hex
	_, err := hex.DecodeString(h)
	require.NoError(t, err, "hash must be hex-encoded")
	require.Equal(t, 64, len(h), "unexpected hash length (expecting SHA-256 hex)")
}

// ========== Benchmarks ==========

var sinkStr string

func BenchmarkMsgHash_Short(b *testing.B) {
	topic := "t"
	msg := []byte("m")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = hash.MsgHash(topic, msg)
	}
	if sinkStr == "" {
		b.Fatal("empty hash")
	}
}

func BenchmarkMsgHash_Large(b *testing.B) {
	topic := "a-very-long-topic-name-to-avoid-inlining"
	msg := bytes.Repeat([]byte("x"), 64*1024) // 64 KiB
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = hash.MsgHash(topic, msg)
	}
	if sinkStr == "" {
		b.Fatal("empty hash")
	}
}

// assert hex ( to catch impl change)
var hexRe = regexp.MustCompile(`\A[0-9a-fA-F]+\z`)

func TestMsgHash_IsHex_WithRegex(t *testing.T) {
	t.Parallel()

	h := hash.MsgHash("topic", []byte("data"))
	require.Truef(t, hexRe.MatchString(h), "hash is not hex: %q", h)
}
