package config_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/internal/endpoints"
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
	cfg.ChainID = "hoodi"
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
		config.WithBootstrapBaseURL(endpoints.DevBootstrapBaseURL),
		config.WithServiceVersion("v0.0.1-rc11"),
		config.WithRenewInterval(2*time.Second),
	)

	// then
	var cfgReceived *entities.DynamicConfig
	select {
	case cfgReceived = <-received:
		require.Equal(t, "hoodi", cfgReceived.ChainID)
		require.Equal(t, "optimum_hoodi_v0_2", cfgReceived.ClusterID)
	case <-time.After(12 * time.Second):
		require.Failf(t, "timeout waiting for config update", "boot node unavailable or config missing for cluster: %s", cfg.ClusterID)
	}
	require.Equal(t, "hoodi", cfgRotator.Get().ChainID)
	require.Equal(t, "optimum_hoodi_v0_2", cfgRotator.Get().ClusterID)
	require.Equal(t, cfgReceived.MeshDegreeMin, cfgRotator.Get().MeshDegreeMin)
	require.Equal(t, cfgReceived.MeshDegreeMax, cfgRotator.Get().MeshDegreeMax)
}

func TestConfigPollUsesServicesToken(t *testing.T) {
	var mu sync.Mutex
	type hit struct {
		path string
		auth string
	}
	var hits []hit
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits = append(hits, hit{path: r.URL.Path, auth: r.Header.Get("Authorization")})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chain_id":"hoodi","cluster_id":"c","propagation_enabled":true}`))
	}))
	t.Cleanup(srv.Close)

	var n atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	config.NewConfigRotator(ctx, logger.NewAppSLogger(logger.Debug), &entities.OptimumConfig{}, "hoodi", "c", nil,
		config.WithBootstrapBaseURL(srv.URL),
		config.WithServiceVersion("v1"),
		config.WithRenewInterval(15*time.Millisecond),
		config.WithAuthToken(func(context.Context) (string, error) {
			id := n.Add(1)
			if id == 1 {
				return "", nil
			}
			return fmt.Sprintf("tok-%d", id), nil
		}),
	)

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		var sawFleet bool
		tokens := map[string]struct{}{}
		for _, h := range hits {
			if h.path == "/api/v1/hoodi/c/config" && h.auth == "" {
				sawFleet = true
			}
			if h.path == "/api/v2/hoodi/c/config" && len(h.auth) > len("Bearer ") {
				tokens[h.auth] = struct{}{}
			}
		}
		return sawFleet && len(tokens) >= 2
	}, 2*time.Second, 10*time.Millisecond)
}

func TestConfigPollSkipsWhenTokenFails(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	config.NewConfigRotator(ctx, logger.NewAppSLogger(logger.Debug), &entities.OptimumConfig{}, "hoodi", "c", nil,
		config.WithBootstrapBaseURL(srv.URL),
		config.WithRenewInterval(10*time.Millisecond),
		config.WithAuthToken(func(context.Context) (string, error) {
			return "", fmt.Errorf("mint failed")
		}),
	)

	time.Sleep(60 * time.Millisecond)
	require.Zero(t, hits.Load())
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
	for range 1000 {
		wg.Go(func() {
			for j := range 10000 {
				cfgRotator.RenewConfig(&entities.DynamicConfig{
					MeshDegreeMin: int64(4 + j%10),
				})
			}
		})
	}
	for range 1000 {
		wg.Go(func() {
			for range 10000 {
				_ = cfgRotator.Get().MeshDegreeMin
			}
		})
	}
	wg.Wait()
}
