package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getoptimum/optimum-common/pkg/hash"
	"gopkg.in/yaml.v3"
)

var (
	supportedChains = map[string]struct{}{
		"eth_mainnet": {},
		"eth_hoodi":   {},

		"default": {},
	}
)

// DynamicConfig represents runtime-configurable parameters for P2P network behavior.
// These settings can be updated without restarting the node.
type DynamicConfig struct {
	ChainID         string    `db:"chain_id" yaml:"chain_id" json:"chain_id"`
	ClusterID       string    `db:"cluster_id" yaml:"cluster_id" json:"cluster_id"`
	UpdatedAt       time.Time `db:"updated_at" yaml:"-" json:"updated_at"`
	EnableABTesting bool      `db:"enable_ab_testing" yaml:"enable_ab_testing" json:"enable_ab_testing"`
	// ExcludeSelfMessages is a flag that indicates whether messages originating from the node itself should be ignored.
	// used for tracking eth latency measurements
	ExcludeSelfMessages bool `db:"exclude_self_messages" yaml:"exclude_self_messages" json:"exclude_self_messages"`

	// RLNC and message settings
	RandomMessageSize        int64   `db:"random_message_size_bytes" yaml:"random_message_size" json:"random_message_size_bytes"`
	ShardFactor              int64   `db:"rlnc_shard_factor" yaml:"shard_factor" json:"rlnc_shard_factor"`
	PublisherShardMultiplier float32 `db:"publisher_shard_multiplier" yaml:"publisher_shard_multiplier" json:"publisher_shard_multiplier"`
	ForwardShardThreshold    float32 `db:"forward_shard_threshold" yaml:"forward_shard_threshold" json:"forward_shard_threshold"`

	// Mesh topology settings
	MeshDegreeTarget int64 `db:"mesh_degree_target" yaml:"mesh_degree_target" json:"mesh_degree_target"`
	MeshDegreeMin    int64 `db:"mesh_degree_min" yaml:"mesh_degree_min" json:"mesh_degree_min"`
	MeshDegreeMax    int64 `db:"mesh_degree_max" yaml:"mesh_degree_max" json:"mesh_degree_max"`
}

// FromYAMLFile loads a DynamicConfig from a YAML file at the specified path.
// Returns an error if the file cannot be opened or parsed.
func FromYAMLFile(path string) (*DynamicConfig, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("error open config file: %w", err)
	}
	defer func() {
		if e := file.Close(); e != nil {
			log.Fatal("Error close config file", e)
		}
	}()

	var cfg DynamicConfig
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decode config file: %w", err)
	}
	return &cfg, nil
}

// Validate checks that all configuration values are within valid ranges.
// Normalizes chainID by converting to lowercase and trimming whitespace.
// Returns an error if validation fails.
func (d *DynamicConfig) Validate() error {
	d.ChainID = strings.ToLower(strings.TrimSpace(d.ChainID))
	if _, ok := supportedChains[d.ChainID]; !ok {
		return fmt.Errorf("unsupported chain_id: %s", d.ChainID)
	}

	if d.RandomMessageSize <= 0 {
		return errors.New("random_message_size must be greater than zero")
	}
	if d.ShardFactor <= 0 {
		return errors.New("shard_factor must be greater than zero")
	}
	if d.PublisherShardMultiplier <= 0 {
		return errors.New("publisher_shard_multiplier must be greater than zero")
	}
	if d.ForwardShardThreshold < 0 || d.ForwardShardThreshold > 1 {
		return errors.New("forward_shard_threshold must be between 0 and 1")
	}
	if d.MeshDegreeMin <= 0 || d.MeshDegreeMax <= 0 || d.MeshDegreeTarget <= 0 {
		return errors.New("mesh degrees must be greater than zero")
	}
	if d.MeshDegreeMin > d.MeshDegreeMax {
		return errors.New("mesh_degree_min cannot be greater than mesh_degree_max")
	}
	if d.MeshDegreeTarget < d.MeshDegreeMin || d.MeshDegreeTarget > d.MeshDegreeMax {
		return errors.New("mesh_degree_target must be between mesh_degree_min and mesh_degree_max")
	}
	return nil
}

// ToMap converts the DynamicConfig to a map representation suitable for storage or serialization.
// The updated_at field is set to the current time.
func (d *DynamicConfig) ToMap() map[string]any {
	return map[string]any{
		"chain_id":              d.ChainID,
		"cluster_id":            d.ClusterID,
		"enable_ab_testing":     d.EnableABTesting,
		"exclude_self_messages": d.ExcludeSelfMessages,
		"updated_at":            time.Now(),

		"random_message_size_bytes":  d.RandomMessageSize,
		"rlnc_shard_factor":          d.ShardFactor,
		"publisher_shard_multiplier": d.PublisherShardMultiplier,
		"forward_shard_threshold":    d.ForwardShardThreshold,

		"mesh_degree_target": d.MeshDegreeTarget,
		"mesh_degree_min":    d.MeshDegreeMin,
		"mesh_degree_max":    d.MeshDegreeMax,
	}
}

// HashRemoteConfig computes a SHA-256 hash of the configurable fields in DynamicConfig.
// Used to detect when remote configuration has changed.
// The hash includes: EnableABTesting, ExcludeSelfMessages, and all RLNC/mesh parameters.
func HashRemoteConfig(cfg *DynamicConfig) string {
	h := sha256.New()
	hash.WriteBool(h, cfg.EnableABTesting)
	hash.WriteInt64(h, cfg.RandomMessageSize)
	hash.WriteInt64(h, cfg.ShardFactor)
	hash.WriteFloat32(h, cfg.PublisherShardMultiplier)
	hash.WriteFloat32(h, cfg.ForwardShardThreshold)
	hash.WriteInt64(h, cfg.MeshDegreeTarget)
	hash.WriteInt64(h, cfg.MeshDegreeMin)
	hash.WriteInt64(h, cfg.MeshDegreeMax)
	hash.WriteBool(h, cfg.ExcludeSelfMessages)
	return hex.EncodeToString(h.Sum(nil))
}
