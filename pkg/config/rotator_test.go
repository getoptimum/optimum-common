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

func init() {
	config.RenewInterval = 2 * time.Second
}

func TestRenewConfig(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	l := logger.NewAppSLogger(logger.Debug)
	var cfg testConfig
	require.NoError(t, config.Load(&cfg))
	cfg.ClusterID = "optimum_hoodi_v0_1"
	cfg.MeshDegreeMin = 1
	cfg.MeshDegreeMax = 1

	received := make(chan *entities.DynamicConfig, 10)
	updater := func(dcCfg *entities.DynamicConfig) {
		received <- dcCfg
	}

	// when
	cfgRotator := config.NewConfigRotator(ctx, l, &cfg.OptimumConfig, cfg.ChainID, cfg.ClusterID, updater)

	// then
	select {
	case cfgReceived := <-received:
		require.Equal(t, "default", cfgReceived.ChainID)
		require.Equal(t, "optimum_hoodi_v0_1", cfgReceived.ClusterID)
	case <-time.After(12 * time.Second):
		t.Error("timeout waiting for config update")
	}
	require.Equal(t, "default", cfgRotator.Get().ChainID)
	require.Equal(t, "optimum_hoodi_v0_1", cfgRotator.Get().ClusterID)
	require.Equal(t, int64(4), cfgRotator.Get().MeshDegreeMin)
	require.Equal(t, int64(12), cfgRotator.Get().MeshDegreeMax)
}

func TestConfigRotatorConcurrentlyTest(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	l := logger.NewAppSLogger(logger.Debug)
	var cfg testConfig
	require.NoError(t, config.Load(&cfg))
	cfg.ClusterID = "optimum_hoodi_v0_1"

	// when
	cfgRotator := config.NewConfigRotator(ctx, l, &cfg.OptimumConfig, cfg.ChainID, cfg.ClusterID, nil)

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
		require.Equal(t, k, config.HashRemoteConfig(v))
	}
}
