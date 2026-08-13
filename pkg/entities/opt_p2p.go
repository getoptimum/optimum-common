package entities

import (
	"errors"
	"fmt"
	"math"
)

// OptimumConfig holds configuration for P2P network settings including
// cluster/chain identifiers, message size limits, sharding parameters,
// mesh topology settings, and bootstrap peer addresses.
type OptimumConfig struct {
	ClusterID      string `yaml:"cluster_id" env:"CLUSTER_ID" flag:"cluster_id"`
	ChainID        string `yaml:"chain_id" env:"CHAIN_ID" flag:"chain_id" default:"default"`
	MaxMessageSize int64  `yaml:"max_message_size_bytes" env:"OPTIMUM_MAX_MSG_SIZE" flag:"max_message_size_bytes" default:"1048576"`

	// Coding and message settings
	RandomMessageSize        uint32  `yaml:"random_message_size_bytes" env:"OPTIMUM_RANDOM_MSG_SIZE" flag:"random_message_size_bytes" default:"512"`
	ShardFactor              uint32  `yaml:"rlnc_shard_factor" env:"OPTIMUM_SHARD_FACTOR" flag:"rlnc_shard_factor" default:"4"`
	PublisherShardMultiplier float64 `yaml:"publisher_shard_multiplier" env:"OPTIMUM_SHARD_MULT" flag:"publisher_shard_multiplier" default:"1.5"`
	ForwardShardThreshold    float64 `yaml:"forward_shard_threshold" env:"OPTIMUM_THRESHOLD" flag:"forward_shard_threshold" default:"0.75"`

	// Mesh topology settings
	MeshDegreeTarget int64 `yaml:"mesh_degree_target" env:"OPTIMUM_MESH_TARGET" flag:"mesh_degree_target" default:"6"`
	MeshDegreeMin    int64 `yaml:"mesh_degree_min" env:"OPTIMUM_MESH_MIN" flag:"mesh_degree_min" default:"4"`
	MeshDegreeMax    int64 `yaml:"mesh_degree_max" env:"OPTIMUM_MESH_MAX" flag:"mesh_degree_max" default:"12"`

	BootstrapPeers []string `yaml:"bootstrap_peers" env:"BOOTSTRAP_PEERS" flag:"bootstrap_peers"`
}

// Validate checks that required fields are set and values are within valid ranges.
// Returns an error if validation fails.
func (cfg *OptimumConfig) Validate() error {
	if cfg.ClusterID == "" {
		return fmt.Errorf("cluster_id is required")
	}
	if cfg.MaxMessageSize <= 0 {
		return fmt.Errorf("max_message_size_bytes must be > 0 and <= %d (got %d)", math.MaxInt64, cfg.MaxMessageSize)
	}
	if cfg.RandomMessageSize <= 0 {
		return errors.New("random_message_size_bytes must be > 0")
	}
	return nil
}

// ApplyDynamicConfig creates a new OptimumConfig by applying dynamic configuration
// overrides to the current config. Returns a new config instance; does not modify the receiver.
func (cfg *OptimumConfig) ApplyDynamicConfig(dcCfg *DynamicConfig) *OptimumConfig {
	newCfg := cfg.Clone()
	newCfg.RandomMessageSize = dcCfg.RandomMessageSize
	newCfg.ShardFactor = dcCfg.ShardFactor
	newCfg.PublisherShardMultiplier = dcCfg.PublisherShardMultiplier
	newCfg.ForwardShardThreshold = dcCfg.ForwardShardThreshold
	newCfg.MeshDegreeTarget = dcCfg.MeshDegreeTarget
	newCfg.MeshDegreeMin = dcCfg.MeshDegreeMin
	newCfg.MeshDegreeMax = dcCfg.MeshDegreeMax
	return newCfg
}

// Clone creates a deep copy of the configuration.
// The BootstrapPeers slice is copied to avoid sharing references.
func (cfg *OptimumConfig) Clone() *OptimumConfig {
	out := *cfg
	if cfg.BootstrapPeers != nil {
		out.BootstrapPeers = append([]string{}, cfg.BootstrapPeers...)
	}
	return &out
}
