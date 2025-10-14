package entities

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/getoptimum/optimum-common/pkg/utils"
)

type P2PMessage struct {
	SourceNodeID   string `json:"source_node_id"`
	UpstreamPeerID string `json:"upstream_peer_id,omitempty"`
	Topic          string `json:"topic,omitempty"`
	Message        []byte `json:"message"`
}

// Marshal serializes the P2PMessage into JSON format.
// for faster processing, we add 2 bytes header and 32 bytes hash of the message content at the beginning
// [1 byte version][1 byte reserved][32 bytes hash of message content][json data...]
// this allows quick filtering of duplicate messages by hash without full JSON parsing
func (m *P2PMessage) Marshal() ([]byte, error) {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	hashMessage := utils.HashSHA256Bytes(m.Message)
	result := make([]byte, 0, 2+len(hashMessage)+len(jsonData))
	result = append(result, 0x01, 0x00)        // version byte + 1 reserved bytes
	result = append(result, hashMessage[:]...) // 32 bytes of hash
	result = append(result, jsonData...)       // rest is JSON data
	return result, nil
}

// UnmarshalP2PMessage deserializes the P2P wire format into a P2PMessage struct.
// the wire format is: [1 byte version][1 byte reserved][32 bytes hash of message content][JSON data...].
// it checks the header, extracts and verifies the hash, and then unmarshals the JSON data.
func UnmarshalP2PMessage(data []byte) (*P2PMessage, error) {
	if len(data) < 34 { // at least version(1) + reserved(1) + hash(32)
		return nil, fmt.Errorf("data too short to be a valid P2PMessage")
	}
	if data[0] != 0x01 {
		return nil, fmt.Errorf("unsupported P2PMessage version: %d", data[0])
	}
	// skip first 2 bytes (version + reserved) and next 32 bytes (hash)
	var msg P2PMessage
	if err := json.Unmarshal(data[34:], &msg); err != nil {
		return nil, err
	}

	// verify hash of the message content
	expectedHash := utils.HashSHA256Bytes(msg.Message)
	actualHash := data[2:34]
	if !bytes.Equal(expectedHash[:], actualHash) {
		return nil, fmt.Errorf("message hash mismatch")
	}
	return &msg, nil
}

// DecodeFrom reads a header-prefixed P2PMessage from an io.Reader and decodes it into the P2PMessage struct.
func DecodeFrom(r io.Reader) (*P2PMessage, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return UnmarshalP2PMessage(data)
}

// MessageIDFn generates or extract a unique message ID based on the content of the P2PMessage.
// if the data is a common.P2PMessage, return the hash of the message content (the 32 bytes after the 2-byte header).
// otherwise, compute the hash of the entire message.
func MessageIDFn(data []byte) string {
	if len(data) > 34 && data[0] == 0x01 { // at least version(1) + reserved(1) + hash(32)
		return hex.EncodeToString(data[2:34])
	}
	return utils.HashSHA256(data)
}
