package logger

import (
	"log/slog"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

// WithAny creates a Field with any value type.
func WithAny(key string, val any) Field {
	return Field{a: slog.Any(key, val)}
}

// WithString creates a Field with a string value.
func WithString(key, v string) Field {
	return Field{a: slog.String(key, v)}
}

// WithBool creates a Field with a boolean value.
func WithBool(key string, v bool) Field {
	return Field{a: slog.Bool(key, v)}
}

// WithInt creates a Field with an int value (platform-dependent size).
func WithInt(key string, v int) Field {
	return Field{a: slog.Int(key, v)} // uses platform int
}

// WithInt32 creates a Field with an int32 value.
func WithInt32(key string, v int32) Field {
	return Field{a: slog.Int64(key, int64(v))}
}

// WithInt64 creates a Field with an int64 value.
func WithInt64(key string, v int64) Field {
	return Field{a: slog.Int64(key, v)}
}

// WithUint creates a Field with a uint value.
// Uses Uint64 internally as slog has no Uint() function.
func WithUint(key string, v uint) Field {
	// slog has no Uint(), but Uint64 works fine and keeps the sign semantics.
	return Field{a: slog.Uint64(key, uint64(v))}
}

// WithUint64 creates a Field with a uint64 value.
func WithUint64(key string, v uint64) Field {
	return Field{a: slog.Uint64(key, v)}
}

// WithFloat32 creates a Field with a float32 value.
func WithFloat32(key string, v float32) Field {
	return Field{a: slog.Float64(key, float64(v))}
}

// WithFloat64 creates a Field with a float64 value.
func WithFloat64(key string, v float64) Field {
	return Field{a: slog.Float64(key, v)}
}

// WithFlow creates a Field with key "flow" and the given value.
func WithFlow(val string) Field {
	return WithString("flow", val)
}

// WithRunID creates a Field with key "run_id" and the given value.
func WithRunID(val string) Field {
	return WithString("run_id", val)
}

// WithModule creates a Field with key "module" and the given value.
// Returns a pointer to the Field.
func WithModule(val string) *Field {
	f := WithString("module", val)
	return &f
}

// WithError creates a Field with key "err" containing the error message.
// Returns a pointer to the Field.
func WithError(err error) *Field {
	f := WithString("err", err.Error())
	return &f
}

// WithFilePath creates a Field with key "file" and the given file path.
func WithFilePath(filePath string) Field {
	return WithString("file", filePath)
}

// WithService creates a Field with key "service" and the given service name.
func WithService(serviceName string) Field {
	return WithString("service", serviceName)
}

// WithPeer adds the peer ID of the given AddrInfo
func WithPeer(info peer.AddrInfo) Field {
	return WithString("peer", info.ID.String())
}

// WithPeerAddrs adds comma separated multiaddrs of the given peer
func WithPeerAddrs(info peer.AddrInfo) Field {
	addrs := make([]string, 0, len(info.Addrs))
	for _, addr := range info.Addrs {
		addrs = append(addrs, addr.String())
	}
	return WithString("peer", strings.Join(addrs, ","))
}

// WithTopic adds a pubsub topic as a string field
func WithTopic(topic string) Field {
	return WithString("topic", topic)
}

// WithTopicBytes adds a topic represented as byte slice
func WithTopicBytes(topic []byte) Field {
	return WithString("topic", string(topic))
}

func WithClusterID(clusterID string) Field {
	return WithString("cluster_id", clusterID)
}

func WithGatewayID(gatewayID string) Field {
	return WithString("gateway_id", gatewayID)
}
