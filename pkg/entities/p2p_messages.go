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

// UnmarshalP2PMessage deserializes JSON data into a P2PMessage struct.
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

// DecodeFrom reads JSON data from an io.Reader and decodes it into the P2PMessage struct.
func (m *P2PMessage) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(m)
}

func MessageIDFn(data []byte) string {
	// if it common.P2PMessage - hash only content of message, it first 32 bytes after 2 bytes header
	// if not - hash whole message
	if len(data) > 34 && data[0] == 0x01 { // at least version(1) + reserved(1) + hash(32)
		return hex.EncodeToString(data[2:34])
	}
	return utils.HashSHA256(data)
}
