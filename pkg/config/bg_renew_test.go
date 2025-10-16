package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/config"
	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/stretchr/testify/require"
)

func TestRenewConfig(t *testing.T) {
	// given
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	received := make(chan *entities.DynamicConfig, 10)
	updater := func(cfg *entities.DynamicConfig) {
		received <- cfg
	}

	// when
	go config.RenewConfig(ctx, 2*time.Second, "default", "optimum_hoodi_v0_1", updater)

	// then
	select {
	case cfg := <-received:
		require.Equal(t, "default", cfg.ChainID)
		require.Equal(t, "optimum_hoodi_v0_1", cfg.ClusterID)
	case <-time.After(12 * time.Second):
		t.Error("timeout waiting for config update")
	}
}
