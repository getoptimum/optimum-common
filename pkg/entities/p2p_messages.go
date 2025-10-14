package entities

type P2PMessage struct {
	SourceNodeID   string `json:"source_node_id"`
	UpstreamPeerID string `json:"upstream_peer_id,omitempty"`
	Topic          string `json:"topic"`
	MessageID      string `json:"message_id"`
	Message        []byte `json:"message"`
}
