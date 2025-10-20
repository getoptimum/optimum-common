package config

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/getoptimum/optimum-common/pkg/utils"
)

const (
	baseURL = "https://bootstrap.getoptimum.io"
)

var (
	RenewInterval = 5 * time.Minute
)

// Rotator holds the default and current configuration and allows atomic updates.
// need dynamically update config in a thread-safe manner for all p2p nodes in the cluster
type Rotator struct {
	// currentOptConfig holds the current active configuration, which may be updated at runtime.
	// use atomic operations to ensure thread-safe access and updates.
	currentOptConfig atomic.Pointer[entities.OptimumConfig]
	updater          func(config *entities.DynamicConfig)
}

func NewConfigRotator(ctx context.Context, baseOptCfg *entities.OptimumConfig, chainID, clusterID string, updater func(config *entities.DynamicConfig)) *Rotator {
	r := &Rotator{
		updater: updater,
	}
	r.currentOptConfig.Store(baseOptCfg)
	go r.bgFetchConfig(ctx, chainID, clusterID)
	return r
}

func (r *Rotator) bgFetchConfig(ctx context.Context, chainID, clusterID string) {
	if chainID == "" || clusterID == "" {
		return // do not fetch if chainID or clusterID is empty
	}
	url := fmt.Sprintf("%s/api/v1/%s/%s/config", baseURL, chainID, clusterID)
	ticker := time.NewTicker(RenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			config, err := fetchRemoteConfig(ctx, url)
			if err != nil {
				continue
			}
			r.RenewConfig(config)
			if r.updater != nil {
				r.updater(config)
			}
		}
	}
}

func (r *Rotator) RenewConfig(cfg *entities.DynamicConfig) {
	currCfg := r.currentOptConfig.Load()
	r.currentOptConfig.Store(currCfg.ApplyDynamicConfig(cfg))
}

func (r *Rotator) Get() *entities.OptimumConfig {
	return r.currentOptConfig.Load()
}

func fetchRemoteConfig(ctx context.Context, url string) (*entities.DynamicConfig, error) {
	res, code, err := utils.GetCurl[entities.DynamicConfig](ctx, url, nil)
	if code == http.StatusOK && res != nil {
		return res, nil
	}
	return nil, fmt.Errorf("failed to fetch remote config, code %d: %w", code, err)
}
