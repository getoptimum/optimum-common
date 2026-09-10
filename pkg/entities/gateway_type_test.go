package entities_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/entities"
)

func TestGatewayTypeFromString(t *testing.T) {
	got, err := entities.GatewayTypeFromString(" Hermes ")
	require.NoError(t, err)
	require.Equal(t, entities.GatewayTypeHermes, got)

	_, err = entities.GatewayTypeFromString("stream")
	require.Error(t, err)

	_, err = entities.GatewayTypeFromString("readonly")
	require.Error(t, err)
}

func TestGatewayTypeCanPublish(t *testing.T) {
	require.True(t, entities.GatewayTypeHermes.CanPublish())
	require.True(t, entities.GatewayTypePartner.CanPublish())
	require.True(t, entities.GatewayTypeRelay.CanPublish())
	require.False(t, entities.GatewayType("").CanPublish())
	require.False(t, entities.GatewayType("forked").CanPublish())
	require.False(t, entities.GatewayType("stream").CanPublish())
}
