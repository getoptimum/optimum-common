package config_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/config"
	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestRenewConfig(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	l := logger.NewAppSLogger(logger.Debug)
	var cfg testConfig
	require.NoError(t, config.Load(&cfg))
	cfg.ClusterID = "optimum_hoodi_v0_2"
	cfg.MeshDegreeMin = 1
	cfg.MeshDegreeMax = 1

	received := make(chan *entities.DynamicConfig, 10)
	updater := func(dcCfg *entities.DynamicConfig) {
		received <- dcCfg
	}

	// when (short interval for test)
	cfgRotator := config.NewConfigRotator(
		ctx,
		l,
		&cfg.OptimumConfig,
		cfg.ChainID,
		cfg.ClusterID,
		updater,
		config.WithServiceVersion("optimum-common-v0.0.1-rc1"),
		config.WithRenewInterval(2*time.Second),
	)

	// then
	var cfgReceived *entities.DynamicConfig
	select {
	case cfgReceived = <-received:
		require.Equal(t, "default", cfgReceived.ChainID)
		require.Equal(t, "optimum_hoodi_v0_2", cfgReceived.ClusterID)
	case <-time.After(12 * time.Second):
		require.Failf(t, "timeout waiting for config update", "boot node unavailable or config missing for cluster: %s", cfg.ClusterID)
	}
	require.Equal(t, "default", cfgRotator.Get().ChainID)
	require.Equal(t, "optimum_hoodi_v0_2", cfgRotator.Get().ClusterID)
	require.Equal(t, cfgReceived.MeshDegreeMin, cfgRotator.Get().MeshDegreeMin)
	require.Equal(t, cfgReceived.MeshDegreeMax, cfgRotator.Get().MeshDegreeMax)
	staticAgg := cfg.AggregationIntervalMs
	if cfgReceived.AggregationIntervalMs != 0 {
		require.Equal(t, cfgReceived.AggregationIntervalMs, cfgRotator.Get().AggregationIntervalMs)
	} else {
		require.Equal(t, staticAgg, cfgRotator.Get().AggregationIntervalMs)
	}
}

func TestConfigRotatorConcurrentlyTest(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	l := logger.NewAppSLogger(logger.Debug)
	var cfg testConfig
	require.NoError(t, config.Load(&cfg))
	cfg.ClusterID = "optimum_hoodi_v0_2"

	// when
	cfgRotator := config.NewConfigRotator(
		ctx,
		l,
		&cfg.OptimumConfig,
		cfg.ChainID,
		cfg.ClusterID,
		nil,
		config.WithServiceVersion("optimum-common-v0.0.1-rc1"),
	)

	// then
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				cfgRotator.RenewConfig(&entities.DynamicConfig{
					MeshDegreeMin: int64(4 + j%10),
				})
			}
		}()
	}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				_ = cfgRotator.Get().MeshDegreeMin
			}
		}()
	}
	wg.Wait()
}
