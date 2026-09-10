package entities_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/entities"
)

func TestGatewayTypeFromString(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want entities.GatewayType
	}{
		{in: "hermes", want: entities.GatewayTypeHermes},
		{in: " partner ", want: entities.GatewayTypePartner},
		{in: "RELAY", want: entities.GatewayTypeRelay},
	} {
		got, err := entities.GatewayTypeFromString(tc.in)
		require.NoError(t, err, tc.in)
		require.Equal(t, tc.want, got, tc.in)
	}

	// stream was the read-only role; it is no longer a mintable type.
	for _, in := range []string{"readonly", "stream", ""} {
		_, err := entities.GatewayTypeFromString(in)
		require.Error(t, err, in)
	}
}

func TestGatewayTypeCanPublish(t *testing.T) {
	require.True(t, entities.GatewayTypeHermes.CanPublish())
	require.True(t, entities.GatewayTypePartner.CanPublish())
	require.True(t, entities.GatewayTypeRelay.CanPublish())
	require.False(t, entities.GatewayType("stream").CanPublish())
	require.False(t, entities.GatewayType("").CanPublish())
	require.False(t, entities.GatewayType("forked").CanPublish())
}
