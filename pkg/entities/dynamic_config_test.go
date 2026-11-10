package entities_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/stretchr/testify/require"
)

func TestDynamicConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		// given
		dc := &entities.DynamicConfig{
			ChainID:                  "default",
			ClusterID:                "test_cluster",
			ServiceVersion:           "v0.0.1",
			RandomMessageSize:        512,
			ShardFactor:              4,
			PublisherShardMultiplier: 3,
			ForwardShardThreshold:    0.75,
			MeshDegreeTarget:         6,
			MeshDegreeMin:            4,
			MeshDegreeMax:            12,
			MaxShardSize:             4096,
			RLNCK:                    16,
			RedundancyFraction:       1.5,
			RelayCreditMultiplier:    5,
			HeartbeatMS:              250,
		}

		// when, then
		require.NoError(t, dc.Validate())
	})
	t.Run("invalid config", func(t *testing.T) {
		// given
		invalidConfigs := []*entities.DynamicConfig{
			{
				ChainID:                  "test",
				ClusterID:                "test_cluster",
				ServiceVersion:           "v0.0.1",
				RandomMessageSize:        512,
				ShardFactor:              4,
				PublisherShardMultiplier: 3,
				ForwardShardThreshold:    0.75,
				MeshDegreeTarget:         6,
				MeshDegreeMin:            4,
				MeshDegreeMax:            12,
			},
			{ChainID: "default", ServiceVersion: "", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 0, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 0, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 0, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: -0.1, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 1.1, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 0, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 0, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 0},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 3, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 13, MeshDegreeMin: 4, MeshDegreeMax: 12},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, MaxShardSize: 8001, RLNCK: 16, RedundancyFraction: 1.5, RelayCreditMultiplier: 5, HeartbeatMS: 250},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, MaxShardSize: 4096, RLNCK: 1, RedundancyFraction: 1.5, RelayCreditMultiplier: 5, HeartbeatMS: 250},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, MaxShardSize: 4096, RLNCK: 16, RedundancyFraction: 0.5, RelayCreditMultiplier: 5, HeartbeatMS: 250},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, MaxShardSize: 4096, RLNCK: 16, RedundancyFraction: 1.5, RelayCreditMultiplier: 11, HeartbeatMS: 250},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, MaxShardSize: 4096, RLNCK: 16, RedundancyFraction: 1.5, RelayCreditMultiplier: 5, HeartbeatMS: 99},
		}

		for _, dc := range invalidConfigs {
			// when, then
			require.Error(t, dc.Validate())
		}
	})
}

func TestDynamicConfig(t *testing.T) {
	// given
	dc := &entities.DynamicConfig{
		ChainID:                  "test",
		ClusterID:                "test_cluster",
		ServiceVersion:           "v0.0.1",
		UpdatedAt:                time.Unix(1763640739, 0),
		PropagationEnabled:       true,
		ExcludeSelfMessages:      true,
		RandomMessageSize:        1,
		ShardFactor:              2,
		PublisherShardMultiplier: 3,
		ForwardShardThreshold:    4,
		MeshDegreeTarget:         5,
		MeshDegreeMin:            6,
		MeshDegreeMax:            7,
		MaxShardSize:             4096,
		RLNCK:                    16,
		RedundancyFraction:       1.5,
		RelayCreditMultiplier:    5,
		HeartbeatMS:              250,
	}

	// when
	res := dc.ToMap()

	// then
	require.Equal(t, map[string]any{
		"chain_id":                   "test",
		"cluster_id":                 "test_cluster",
		"service_version":            "v0.0.1",
		"propagation_enabled":        true,
		"exclude_self_messages":      true,
		"updated_at":                 res["updated_at"],
		"random_message_size_bytes":  uint32(1),
		"rlnc_shard_factor":          uint32(2),
		"publisher_shard_multiplier": float64(3),
		"forward_shard_threshold":    float64(4),
		"mesh_degree_target":         int64(5),
		"mesh_degree_min":            int64(6),
		"mesh_degree_max":            int64(7),
		"max_shard_size":             uint32(4096),
		"rlnc_k":                     uint32(16),
		"redundancy_fraction":        float64(1.5),
		"relay_credit_multiplier":    5,
		"heartbeat_ms":               250,
	}, res)
}

func TestHashRemoteConfig(t *testing.T) {
	// given
	table := map[string]*entities.DynamicConfig{
		"10cc3c382b13ad9246b74708d03528d294522c558727bd2ed4a242bfb7cf0c3f": {},
		"3b8cf7a5c4c10bde0f34acc83adb35e6846069352d1259888beb0e4e80d4d061": {
			RandomMessageSize:        1,
			ShardFactor:              1,
			PublisherShardMultiplier: 1,
			ForwardShardThreshold:    1,
			MeshDegreeTarget:         1,
			MeshDegreeMin:            1,
			MeshDegreeMax:            1,
		},
		"e5cbb441185f89bfd5ed0856fd041ef26d75d9b84945e2c738bd7e6cf89c887b": {
			RandomMessageSize:        2,
			ShardFactor:              2,
			PublisherShardMultiplier: 2,
			ForwardShardThreshold:    2,
			MeshDegreeTarget:         2,
			MeshDegreeMin:            2,
			MeshDegreeMax:            2,
		},
		"e540ddb3093bf748ee6f75672dfcc68820ac98876a73e1cc87d0044b8eb6fa4d": {
			RandomMessageSize:        2,
			ShardFactor:              2,
			PublisherShardMultiplier: 2,
			ForwardShardThreshold:    2,
			MeshDegreeTarget:         2,
			MeshDegreeMin:            2,
			MeshDegreeMax:            2,
			ExcludeSelfMessages:      true,
		},
	}
	for k, v := range table {
		require.Equal(t, k, entities.HashRemoteConfig(v))
	}
}

func TestDynamicConfig_PropagationEnabled(t *testing.T) {
	cfg := &entities.DynamicConfig{
		ChainID:            "hoodi",
		ClusterID:          "c1",
		ServiceVersion:     "v1",
		PropagationEnabled: true,
	}

	m := cfg.ToMap()
	require.Equal(t, true, m["propagation_enabled"])

	// JSON round-trips under the positive key.
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.Contains(t, string(b), `"propagation_enabled":true`)

	var back entities.DynamicConfig
	require.NoError(t, json.Unmarshal([]byte(`{"propagation_enabled":true}`), &back))
	require.True(t, back.PropagationEnabled)

	// PropagationEnabled is now part of the config hash.
	require.NotEqual(t,
		entities.HashRemoteConfig(&entities.DynamicConfig{PropagationEnabled: false}),
		entities.HashRemoteConfig(&entities.DynamicConfig{PropagationEnabled: true}),
	)
}
