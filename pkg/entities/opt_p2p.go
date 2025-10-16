package entities

import (
	"fmt"
	"math"
)

var _ DCConfigurable = (*OptimumConfig)(nil)

type OptimumConfig struct {
	ClusterID      string `yaml:"cluster_id" env:"CLUSTER_ID" flag:"cluster_id"`
	MaxMessageSize int64  `yaml:"max_message_size_bytes" env:"OPTIMUM_MAX_MSG_SIZE" flag:"max_message_size_bytes" default:"1048576"`

	// RLNC and message settings
	RandomMessageSize        int64   `yaml:"random_message_size_bytes" env:"OPTIMUM_RANDOM_MSG_SIZE" flag:"random_message_size_bytes" default:"512"`
	ShardFactor              int64   `yaml:"rlnc_shard_factor" env:"OPTIMUM_SHARD_FACTOR" flag:"rlnc_shard_factor" default:"4"`
	PublisherShardMultiplier float32 `yaml:"publisher_shard_multiplier" env:"OPTIMUM_SHARD_MULT" flag:"publisher_shard_multiplier" default:"1.5"`
	ForwardShardThreshold    float32 `yaml:"forward_shard_threshold" env:"OPTIMUM_THRESHOLD" flag:"forward_shard_threshold" default:"0.75"`

	// Mesh topology settings
	MeshDegreeTarget int64 `yaml:"mesh_degree_target" env:"OPTIMUM_MESH_TARGET" flag:"mesh_degree_target" default:"6"`
	MeshDegreeMin    int64 `yaml:"mesh_degree_min" env:"OPTIMUM_MESH_MIN" flag:"mesh_degree_min" default:"4"`
	MeshDegreeMax    int64 `yaml:"mesh_degree_max" env:"OPTIMUM_MESH_MAX" flag:"mesh_degree_max" default:"12"`

	BootstrapPeers []string `yaml:"bootstrap_peers" env:"BOOTSTRAP_PEERS" flag:"bootstrap_peers"`
}

func (cfg *OptimumConfig) Validate() error {
	if cfg.ClusterID == "" {
		return fmt.Errorf("cluster_id is required")
	}
	if cfg.MaxMessageSize <= 0 || cfg.MaxMessageSize > math.MaxInt {
		return fmt.Errorf("random message size must be positive and less than max int: %d", cfg.MaxMessageSize)
	}
	return nil
}

func (cfg *OptimumConfig) ApplyDynamicConfig(dcCfg *DynamicConfig) {
	cfg.RandomMessageSize = dcCfg.RandomMessageSize
	cfg.ShardFactor = dcCfg.ShardFactor
	cfg.PublisherShardMultiplier = dcCfg.PublisherShardMultiplier
	cfg.ForwardShardThreshold = dcCfg.ForwardShardThreshold
	cfg.MeshDegreeTarget = dcCfg.MeshDegreeTarget
	cfg.MeshDegreeMin = dcCfg.MeshDegreeMin
	cfg.MeshDegreeMax = dcCfg.MeshDegreeMax
}
