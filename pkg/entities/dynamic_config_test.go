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
		"random_message_size_bytes":  int64(1),
		"rlnc_shard_factor":          int64(2),
		"publisher_shard_multiplier": float32(3),
		"forward_shard_threshold":    float32(4),
		"mesh_degree_target":         int64(5),
		"mesh_degree_min":            int64(6),
		"mesh_degree_max":            int64(7),
	}, res)
}

func TestHashRemoteConfig(t *testing.T) {
	// given
	table := map[string]*entities.DynamicConfig{
		"cc2786e1f9910a9d811400edcddaf7075195f7a16b216dcbefba3bc7c4f2ae51": {},
		"db60d9b58747f20f922f9660d71631d4b26465836096857cd53eca21883dbad8": {
			RandomMessageSize:        1,
			ShardFactor:              1,
			PublisherShardMultiplier: 1,
			ForwardShardThreshold:    1,
			MeshDegreeTarget:         1,
			MeshDegreeMin:            1,
			MeshDegreeMax:            1,
		},
		"6a6daffb5ebc958a231ab5099b2d935bf5916e3e4eeab44f8f9c70016d72b672": {
			RandomMessageSize:        2,
			ShardFactor:              2,
			PublisherShardMultiplier: 2,
			ForwardShardThreshold:    2,
			MeshDegreeTarget:         2,
			MeshDegreeMin:            2,
			MeshDegreeMax:            2,
		},
		"f7dbb42586df129bfc7cadcaf952db8256c174915e4ac301cba67a3a2144153e": {
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
		// when, then
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
