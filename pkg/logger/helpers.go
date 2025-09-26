package logger

import (
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

func WithAny(key string, val any) Field {
	return Field{key: key, value: val}
}

func WithString(key, val string) Field {
	return Field{key: key, value: val}
}

func WithUint64(key string, val uint64) Field {
	return Field{key: key, value: val}
}

func WithBool(key string, val bool) Field {
	return Field{key: key, value: val}
}

func WithInt64(key string, val int64) Field {
	return Field{key: key, value: val}
}

func WithInt(key string, val int) Field {
	return Field{key: key, value: val}
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
