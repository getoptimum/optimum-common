package logger

import (
	"log/slog"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

func WithAny(key string, val any) Field {
	return Field{a: slog.Any(key, val)}
}

func WithString(key, v string) Field {
	return Field{a: slog.String(key, v)}
}

func WithBool(key string, v bool) Field {
	return Field{a: slog.Bool(key, v)}
}

func WithInt(key string, v int) Field {
	return Field{a: slog.Int(key, v)} // uses platform int
}

func WithInt32(key string, v int32) Field {
	return Field{a: slog.Int64(key, int64(v))}
}

func WithInt64(key string, v int64) Field {
	return Field{a: slog.Int64(key, v)}
}

func WithUint(key string, v uint) Field {
	// slog has no Uint(), but Uint64 works fine and keeps the sign semantics.
	return Field{a: slog.Uint64(key, uint64(v))}
}

func WithUint64(key string, v uint64) Field {
	return Field{a: slog.Uint64(key, v)}
}

func WithFloat32(key string, v float32) Field {
	return Field{a: slog.Float64(key, float64(v))}
}

func WithFloat64(key string, v float64) Field {
	return Field{a: slog.Float64(key, v)}
}

func WithFlow(val string) Field {
	return WithString("flow", val)
}

func WithRunID(val string) Field {
	return WithString("run_id", val)
}

func WithModule(val string) *Field {
	f := WithString("module", val)
	return &f
}

func WithError(err error) *Field {
	f := WithString("err", err.Error())
	return &f
}

func WithFilePath(filePath string) Field {
	return WithString("file", filePath)
}

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
