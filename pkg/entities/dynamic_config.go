package entities

import "time"

type DynamicConfig struct {
	ChainID         string    `json:"chain_id"`
	ClusterID       string    `json:"cluster_id"`
	UpdatedAt       time.Time `json:"updated_at"`
	EnableABTesting bool      `json:"enable_ab_testing"`
	// SkipMessageFromSelf flag that indicates whether messages originating from the node itself should be ignored.
	// used for track eth latency measurements
	SkipMessageFromSelf bool `json:"skip_message_from_self"`

	// RLNC and message settings
	RandomMessageSize        int64   `json:"random_message_size_bytes"`
	ShardFactor              int64   `json:"rlnc_shard_factor"`
	PublisherShardMultiplier float32 `json:"publisher_shard_multiplier"`
	ForwardShardThreshold    float32 `json:"forward_shard_threshold"`

	// Mesh topology settings
	MeshDegreeTarget int64 `json:"mesh_degree_target"`
	MeshDegreeMin    int64 `json:"mesh_degree_min"`
	MeshDegreeMax    int64 `json:"mesh_degree_max"`
}
