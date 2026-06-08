package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/entities"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestTokenAudienceString(t *testing.T) {
	require.Equal(t, "p2p", entities.TokenAudienceP2P.String())
	require.Equal(t, "services", entities.TokenAudienceServices.String())
}

func TestGatewayClaimsHasAudience(t *testing.T) {
	c := entities.GatewayClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience: jwt.ClaimStrings{entities.TokenAudienceServices.String()},
		},
	}
	require.True(t, c.HasAudience(entities.TokenAudienceServices))
	require.False(t, c.HasAudience(entities.TokenAudienceP2P))

	// Empty aud matches nothing.
	require.False(t, (&entities.GatewayClaims{}).HasAudience(entities.TokenAudienceP2P))
}

func TestGatewayClaimsJSON(t *testing.T) {
	// services-token shape: operator_id present, cnf present.
	services := entities.GatewayClaims{
		ScopeVersion: 1,
		Type:         entities.GatewayTypePartner,
		ChainID:      "hoodi",
		OperatorID:   "42",
		CNF:          entities.GatewayConfirmation{PeerID: "12D3KooWpeer"},
	}
	raw, err := json.Marshal(services)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"operator_id":"42"`)
	require.Contains(t, string(raw), `"peer_id":"12D3KooWpeer"`)

	// handshake-token shape: operator_id omitted (omitempty) so it never leaks
	// onto the peer-visible token.
	handshake := entities.GatewayClaims{
		ScopeVersion: 1,
		Type:         entities.GatewayTypePartner,
		ChainID:      "hoodi",
	}
	raw, err = json.Marshal(handshake)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "operator_id")
}
