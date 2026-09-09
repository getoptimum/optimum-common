package entities_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/stretchr/testify/require"
)

func TestGatewayTypeFromString(t *testing.T) {
	got, err := entities.GatewayTypeFromString(" Stream ")
	require.NoError(t, err)
	require.Equal(t, entities.GatewayTypeStream, got)

	_, err = entities.GatewayTypeFromString("readonly")
	require.Error(t, err)
}

func TestGatewayTypeCanPublish(t *testing.T) {
	require.True(t, entities.GatewayTypeHermes.CanPublish())
	require.True(t, entities.GatewayTypePartner.CanPublish())
	require.True(t, entities.GatewayTypeRelay.CanPublish())
	require.False(t, entities.GatewayTypeStream.CanPublish())
	require.False(t, entities.GatewayType("").CanPublish())
	require.False(t, entities.GatewayType("forked").CanPublish())
}
