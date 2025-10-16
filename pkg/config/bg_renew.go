package config

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/getoptimum/optimum-common/pkg/utils"
)

const (
	baseURL = "https://bootstrap.getoptimum.io"
)

func RenewConfig(ctx context.Context, renewInterval time.Duration, chainID, clusterID string, updater func(config *entities.DynamicConfig)) {
	url := fmt.Sprintf("%s/api/v1/%s/%s/config", baseURL, chainID, clusterID)
	ticker := time.NewTicker(renewInterval)
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
			updater(config)
		}
	}
}

func fetchRemoteConfig(ctx context.Context, url string) (*entities.DynamicConfig, error) {
	res, code, err := utils.GetCurl[entities.DynamicConfig](ctx, url, nil)
	if code == http.StatusOK && res != nil {
		return res, nil
	}
	return nil, fmt.Errorf("failed to fetch remote config, code %d: %w", code, err)
}
