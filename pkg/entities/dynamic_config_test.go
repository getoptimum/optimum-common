package entities_test

import (
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
	t.Run("valid aggregation intervals", func(t *testing.T) {
		for _, ms := range []int64{300000, 600000} {
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
				AggregationIntervalMs:    ms,
			}
			require.NoError(t, dc.Validate(), "ms=%d", ms)
		}
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
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, AggregationIntervalMs: 600001},
			{ChainID: "default", ServiceVersion: "v0.0.1", RandomMessageSize: 512, ShardFactor: 4, PublisherShardMultiplier: 3, ForwardShardThreshold: 0.75, MeshDegreeTarget: 6, MeshDegreeMin: 4, MeshDegreeMax: 12, AggregationIntervalMs: -1},
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
		EnableABTesting:          true,
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
		"enable_ab_testing":          true,
		"exclude_self_messages":      true,
		"updated_at":                 res["updated_at"],
		"random_message_size_bytes":  int64(1),
		"rlnc_shard_factor":          int64(2),
		"publisher_shard_multiplier": float32(3),
		"forward_shard_threshold":    float32(4),
		"mesh_degree_target":         int64(5),
		"mesh_degree_min":            int64(6),
		"mesh_degree_max":            int64(7),
		"aggregation_interval_ms":    int64(0),
	}, res)
}

func TestHashRemoteConfig(t *testing.T) {
	// given
	table := map[string]*entities.DynamicConfig{
		"66b4a8b2a17f0463f7427c0239106eaf710ea7129f42d184a58c50cdff614ba4": {},
		"976516c521f1eb41ddaa7d95f97492a3e2cac80bfe3da6682398d65bc19882e5": {
			RandomMessageSize:        1,
			ShardFactor:              1,
			PublisherShardMultiplier: 1,
			ForwardShardThreshold:    1,
			MeshDegreeTarget:         1,
			MeshDegreeMin:            1,
			MeshDegreeMax:            1,
		},
		"70ff9ec91b49a0b86f868fd1a2c1ac6d8fd4d93d55419cb89dd7266b211b7c3a": {
			RandomMessageSize:        2,
			ShardFactor:              2,
			PublisherShardMultiplier: 2,
			ForwardShardThreshold:    2,
			MeshDegreeTarget:         2,
			MeshDegreeMin:            2,
			MeshDegreeMax:            2,
		},
		"47909cd599f3600575ccdba3306caf4730a2ed11cb0a2aa3e5505a8fec1f62f6": {
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
