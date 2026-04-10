package entities_test

import (
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/config"
	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/stretchr/testify/require"
)

func TestOptimumConfig(t *testing.T) {
	var cfg entities.OptimumConfig
	require.NoError(t, config.Load(&cfg))

	t.Run("verify defaults", func(t *testing.T) {
		require.Error(t, cfg.Validate())
		cfg.ClusterID = "test-cluster"
		require.NoError(t, cfg.Validate())
		require.Equal(t, "test-cluster", cfg.ClusterID)
		require.Equal(t, int64(1048576), cfg.MaxMessageSize)
		require.Equal(t, int64(512), cfg.RandomMessageSize)
		require.Equal(t, int64(4), cfg.ShardFactor)
		require.Equal(t, float32(1.5), cfg.PublisherShardMultiplier)
		require.Equal(t, float32(0.75), cfg.ForwardShardThreshold)
		require.Equal(t, int64(6), cfg.MeshDegreeTarget)
		require.Equal(t, int64(4), cfg.MeshDegreeMin)
		require.Equal(t, int64(12), cfg.MeshDegreeMax)
	})
	t.Run("aggregation_interval_ms boundaries on OptimumConfig.Validate", func(t *testing.T) {
		base := entities.OptimumConfig{
			ClusterID:         "test-cluster",
			MaxMessageSize:    1048576,
			RandomMessageSize: 512,
		}
		t.Run("zero uses default and is valid", func(t *testing.T) {
			c := base
			c.AggregationIntervalMs = 0
			require.NoError(t, c.Validate())
		})
		t.Run("max interval is valid", func(t *testing.T) {
			c := base
			c.AggregationIntervalMs = entities.MaxAggregationIntervalMs
			require.NoError(t, c.Validate())
		})
		t.Run("above max is invalid", func(t *testing.T) {
			c := base
			c.AggregationIntervalMs = entities.MaxAggregationIntervalMs + 1
			require.Error(t, c.Validate())
		})
		t.Run("negative is invalid", func(t *testing.T) {
			c := base
			c.AggregationIntervalMs = -1
			require.Error(t, c.Validate())
		})
	})
	t.Run("should apply dynamic config", func(t *testing.T) {
		dc := &entities.DynamicConfig{
			ChainID:                  "default",
			ClusterID:                "dynamic-cluster",
			UpdatedAt:                time.Now(),
			EnableABTesting:          true,
			RandomMessageSize:        1024,
			ShardFactor:              8,
			PublisherShardMultiplier: 3,
			ForwardShardThreshold:    1.23,
			MeshDegreeTarget:         32,
			MeshDegreeMin:            64,
			MeshDegreeMax:            128,
		}

		// when
		newCfg := cfg.ApplyDynamicConfig(dc)

		// then
		require.Equal(t, int64(1024), newCfg.RandomMessageSize)
		require.Equal(t, int64(8), newCfg.ShardFactor)
		require.Equal(t, float32(3), newCfg.PublisherShardMultiplier)
		require.Equal(t, float32(1.23), newCfg.ForwardShardThreshold)
		require.Equal(t, int64(32), newCfg.MeshDegreeTarget)
		require.Equal(t, int64(64), newCfg.MeshDegreeMin)
		require.Equal(t, int64(128), newCfg.MeshDegreeMax)
	})
}
