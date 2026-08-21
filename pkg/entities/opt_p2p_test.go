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
		require.Equal(t, uint32(512), cfg.RandomMessageSize)
		require.Equal(t, uint32(4), cfg.ShardFactor)
		require.Equal(t, float64(1.5), cfg.PublisherShardMultiplier)
		require.Equal(t, float64(0.75), cfg.ForwardShardThreshold)
		require.Equal(t, int64(6), cfg.MeshDegreeTarget)
		require.Equal(t, int64(4), cfg.MeshDegreeMin)
		require.Equal(t, int64(12), cfg.MeshDegreeMax)
	})
	t.Run("should apply dynamic config", func(t *testing.T) {
		dc := &entities.DynamicConfig{
			ChainID:                  "default",
			ClusterID:                "dynamic-cluster",
			UpdatedAt:                time.Now(),
			PropagationEnabled:       false,
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
		require.Equal(t, uint32(1024), newCfg.RandomMessageSize)
		require.Equal(t, uint32(8), newCfg.ShardFactor)
		require.Equal(t, float64(3), newCfg.PublisherShardMultiplier)
		require.Equal(t, 1.23, newCfg.ForwardShardThreshold)
		require.Equal(t, int64(32), newCfg.MeshDegreeTarget)
		require.Equal(t, int64(64), newCfg.MeshDegreeMin)
		require.Equal(t, int64(128), newCfg.MeshDegreeMax)
	})
}
