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
type Rotator[T any] struct {
	// currentConfig holds the current active configuration, which may be updated at runtime.
	// use atomic operations to ensure thread-safe access and updates.
	currentConfig atomic.Pointer[T]

	updater func(config *entities.DynamicConfig) *T
}

func NewConfigRotator[T any](ctx context.Context, cfg *T, chainID, clusterID string, updater func(config *entities.DynamicConfig) *T) *Rotator[T] {
	r := &Rotator[T]{
		currentConfig: atomic.Pointer[T]{},
		updater:       updater,
	}
	r.currentConfig.Store(cfg)
	go r.bgFetchConfig(ctx, chainID, clusterID)
	return r
}

func (r *Rotator[T]) bgFetchConfig(ctx context.Context, chainID, clusterID string) {
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
		}
	}
}

func (r *Rotator[T]) RenewConfig(cfg *entities.DynamicConfig) {
	r.currentConfig.Store(r.updater(cfg))
}

func (r *Rotator[T]) Get() *T {
	return r.currentConfig.Load()
}

func fetchRemoteConfig(ctx context.Context, url string) (*entities.DynamicConfig, error) {
	res, code, err := utils.GetCurl[entities.DynamicConfig](ctx, url, nil)
	if code == http.StatusOK && res != nil {
		return res, nil
	}
	return nil, fmt.Errorf("failed to fetch remote config, code %d: %w", code, err)
}
