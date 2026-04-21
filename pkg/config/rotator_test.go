package config_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/config"
	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/stretchr/testify/require"
)

func newFakeBootstrap(t *testing.T, response *entities.DynamicConfig) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRenewConfig(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	l := logger.NewAppSLogger(logger.Debug)

	fakeCfg := entities.DynamicConfig{
		ChainID:       "test-chain",
		ClusterID:     "test-cluster",
		MeshDegreeMin: 2,
		MeshDegreeMax: 8,
	}
	srv := newFakeBootstrap(t, &fakeCfg)

	var cfg testConfig
	require.NoError(t, config.Load(&cfg))
	cfg.ClusterID = fakeCfg.ClusterID
	cfg.ChainID = fakeCfg.ChainID
	cfg.MeshDegreeMin = 1
	cfg.MeshDegreeMax = 1

	received := make(chan *entities.DynamicConfig, 10)
	updater := func(dcCfg *entities.DynamicConfig) {
		received <- dcCfg
	}

	// when (fast interval; test server responds immediately)
	cfgRotator := config.NewConfigRotator(
		ctx,
		l,
		&cfg.OptimumConfig,
		cfg.ChainID,
		cfg.ClusterID,
		updater,
		config.WithServiceVersion("optimum-common-v0.0.1-rc1"),
		config.WithRenewInterval(100*time.Millisecond),
		config.WithBaseURL(srv.URL),
	)

	// then
	var cfgReceived *entities.DynamicConfig
	select {
	case cfgReceived = <-received:
		require.Equal(t, fakeCfg.ChainID, cfgReceived.ChainID)
		require.Equal(t, fakeCfg.ClusterID, cfgReceived.ClusterID)
	case <-time.After(2 * time.Second):
		require.Failf(t, "timeout waiting for config update",
			"no config received within timeout — check that chain %q / cluster %q resolves at %s",
			cfg.ChainID, cfg.ClusterID, srv.URL)
	}
	require.Equal(t, fakeCfg.ChainID, cfgRotator.Get().ChainID)
	require.Equal(t, fakeCfg.ClusterID, cfgRotator.Get().ClusterID)
	require.Equal(t, cfgReceived.MeshDegreeMin, cfgRotator.Get().MeshDegreeMin)
	require.Equal(t, cfgReceived.MeshDegreeMax, cfgRotator.Get().MeshDegreeMax)
}

func TestWithHTTPClient(t *testing.T) {
	// given — a RoundTripper that records whether it was invoked
	called := make(chan struct{}, 1)
	fakeCfg := entities.DynamicConfig{
		ChainID:   "test-chain",
		ClusterID: "test-cluster",
	}
	srv := newFakeBootstrap(t, &fakeCfg)

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called <- struct{}{}
		return http.DefaultTransport.RoundTrip(req)
	})
	customClient := &http.Client{Transport: rt}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	l := logger.NewAppSLogger(logger.Debug)
	var cfg testConfig
	require.NoError(t, config.Load(&cfg))

	received := make(chan *entities.DynamicConfig, 1)
	config.NewConfigRotator(
		ctx,
		l,
		&cfg.OptimumConfig,
		"test-chain",
		"test-cluster",
		func(dc *entities.DynamicConfig) { received <- dc },
		config.WithBaseURL(srv.URL),
		config.WithHTTPClient(customClient),
		config.WithRenewInterval(100*time.Millisecond),
	)

	// then — the injected client must be used for the fetch
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("custom HTTP client was not called within timeout")
	}
	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("no config received via custom HTTP client within timeout")
	}
}

// roundTripperFunc allows using a plain function as an http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestConfigRotatorConcurrentlyTest(t *testing.T) {
	// given — empty chain/cluster disables the background fetch; this test only
	// exercises concurrent RenewConfig / Get safety, not network behavior.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	l := logger.NewAppSLogger(logger.Debug)
	var cfg testConfig
	require.NoError(t, config.Load(&cfg))

	// when
	cfgRotator := config.NewConfigRotator(
		ctx,
		l,
		&cfg.OptimumConfig,
		"", // empty disables bgFetchConfig
		"",
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
