package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestHelperFields(t *testing.T) {
	t.Parallel()

	addr1, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
	require.NoError(t, err)
	addr2, err := multiaddr.NewMultiaddr("/ip6/::1/udp/9000/quic")
	require.NoError(t, err)

	peerID := peer.ID("12D3KooWRy5nTrg1TeStPeerIDabcdefgh")
	peerInfo := peer.AddrInfo{ID: peerID}
	peerInfoWithAddrs := peer.AddrInfo{Addrs: []multiaddr.Multiaddr{addr1, addr2}}

	sampleErr := errors.New("helper failure")

	tests := []struct {
		name     string
		field    logger.Field
		key      string
		expected any
	}{
		{
			name:     "WithAny",
			field:    logger.WithAny("any", "value"),
			key:      "any",
			expected: "value",
		},
		{
			name:     "WithString",
			field:    logger.WithString("string", "text"),
			key:      "string",
			expected: "text",
		},
		{
			name:     "WithUint64",
			field:    logger.WithUint64("u64", 42),
			key:      "u64",
			expected: float64(42),
		},
		{
			name:     "WithBool",
			field:    logger.WithBool("bool", true),
			key:      "bool",
			expected: true,
		},
		{
			name:     "WithInt64",
			field:    logger.WithInt64("i64", -11),
			key:      "i64",
			expected: float64(-11),
		},
		{
			name:     "WithInt",
			field:    logger.WithInt("int", 17),
			key:      "int",
			expected: float64(17),
		},
		{
			name:     "WithModule",
			field:    *logger.WithModule("module-a"),
			key:      "module",
			expected: "module-a",
		},
		{
			name:     "WithError",
			field:    *logger.WithError(sampleErr),
			key:      "err",
			expected: sampleErr.Error(),
		},
		{
			name:     "WithFilePath",
			field:    logger.WithFilePath("/tmp/file"),
			key:      "file",
			expected: "/tmp/file",
		},
		{
			name:     "WithService",
			field:    logger.WithService("api"),
			key:      "service",
			expected: "api",
		},
		{
			name:     "WithPeer",
			field:    logger.WithPeer(peerInfo),
			key:      "peer",
			expected: peerID.String(),
		},
		{
			name:     "WithPeerAddrs",
			field:    logger.WithPeerAddrs(peerInfoWithAddrs),
			key:      "peer",
			expected: addr1.String() + "," + addr2.String(),
		},
		{
			name:     "WithTopic",
			field:    logger.WithTopic("topic-name"),
			key:      "topic",
			expected: "topic-name",
		},
		{
			name:     "WithTopicBytes",
			field:    logger.WithTopicBytes([]byte("bytes-topic")),
			key:      "topic",
			expected: "bytes-topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			value := logFieldValue(t, tt.field, tt.key)
			require.Equal(t, tt.expected, value)
		})
	}
}

func logFieldValue(t *testing.T, field logger.Field, key string) any {
	t.Helper()

	var buf bytes.Buffer
	writers := []io.Writer{&buf}
	l := logger.InitLogger(writers, "test", logger.Debug)
	l.Info("message", field)

	raw := bytes.TrimSpace(buf.Bytes())
	require.NotEmpty(t, raw)

	entry := make(map[string]any)
	require.NoError(t, json.Unmarshal(raw, &entry))

	value, ok := entry[key]
	require.Truef(t, ok, "expected key %q in log entry", key)

	return value
}
