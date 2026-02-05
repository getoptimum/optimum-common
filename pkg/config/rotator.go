package config

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/getoptimum/optimum-common/pkg/logger"
	netpkg "github.com/getoptimum/optimum-common/pkg/net"
)

const (
	baseURL = "https://bootstrap.getoptimum.io"
)

var (
	RenewInterval = 1 * time.Minute
)

// RotatorOption configures a Rotator (e.g. WithRenewInterval).
type RotatorOption func(*Rotator)

// WithRenewInterval sets the interval between config fetch attempts. If not set, RenewInterval is used.
func WithRenewInterval(d time.Duration) RotatorOption {
	return func(r *Rotator) { r.renewInterval = d }
}

// Rotator holds the default and current configuration and allows atomic updates.
// need dynamically update config in a thread-safe manner for all p2p nodes in the cluster
type Rotator struct {
	log              logger.AppLogger
	renewInterval    time.Duration
	lastHash         string
	currentOptConfig atomic.Pointer[entities.OptimumConfig]
	updater          func(config *entities.DynamicConfig)
}

// NewConfigRotator creates a new config rotator that periodically fetches and applies
// dynamic configuration from a remote endpoint.
// The rotator starts a background goroutine that fetches config at the specified interval.
// If chainID or clusterID is empty, fetching is disabled.
// The updater function is called whenever a new config is successfully fetched and applied.
func NewConfigRotator(
	ctx context.Context,
	log logger.AppLogger,
	baseOptCfg *entities.OptimumConfig,
	chainID,
	clusterID string,
	updater func(config *entities.DynamicConfig),
	opts ...RotatorOption,
) *Rotator {
	r := &Rotator{
		log:     log.With(logger.WithService("config_rotator")),
		updater: updater,
	}
	for _, o := range opts {
		o(r)
	}
	r.currentOptConfig.Store(baseOptCfg.Clone())
	go r.bgFetchConfig(ctx, chainID, clusterID)
	return r
}

func (r *Rotator) bgFetchConfig(ctx context.Context, chainID, clusterID string) {
	if chainID == "" || clusterID == "" {
		return // do not fetch if chainID or clusterID is empty
	}
	url := fmt.Sprintf("%s/api/v1/%s/%s/config", baseURL, chainID, clusterID)
	config, err := fetchRemoteConfig(ctx, url)
	if err == nil { // if init failed, we not panic, use default one, just try later fetch it again
		r.RenewConfig(config)
		if r.updater != nil {
			r.updater(config)
		}
	}

	interval := r.renewInterval
	if interval == 0 {
		interval = RenewInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			config, err = fetchRemoteConfig(ctx, url)
			if err != nil {
				continue
			}
			newObjHash := entities.HashRemoteConfig(config)
			if r.lastHash == newObjHash {
				continue
			}
			r.log.Info("applying new dynamic config",
				logger.WithString("old_hash", r.lastHash),
				logger.WithString("new_hash", newObjHash),
			)
			r.lastHash = newObjHash

			r.RenewConfig(config)
			if r.updater != nil {
				r.updater(config)
			}
		}
	}
}

// RenewConfig atomically updates the current configuration by applying the dynamic config.
// Thread-safe: uses atomic operations for concurrent access.
func (r *Rotator) RenewConfig(cfg *entities.DynamicConfig) {
	currCfg := r.currentOptConfig.Load()
	r.currentOptConfig.Store(currCfg.ApplyDynamicConfig(cfg))
}

// Get returns the current configuration.
// Thread-safe: uses atomic operations for concurrent access.
func (r *Rotator) Get() *entities.OptimumConfig {
	return r.currentOptConfig.Load()
}

func fetchRemoteConfig(ctx context.Context, url string) (*entities.DynamicConfig, error) {
	res, code, err := netpkg.GetCurl[entities.DynamicConfig](ctx, url, nil)
	if code == http.StatusOK && res != nil {
		return res, nil
	}
	return nil, fmt.Errorf("failed to fetch remote config, code %d: %w", code, err)
}
