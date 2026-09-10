package entities_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/entities"
)

// The wire values optimum-auth emits. A typo in either silently changes what is
// granted, and the mismatch would only surface against a live token.
func TestGrantConstants(t *testing.T) {
	require.Equal(t, "p2p:publish", entities.GrantP2PPublish)
	require.Equal(t, "p2p:subscribe", entities.GrantP2PSubscribe)
}

// Grants decide publish whenever the claim states any, whatever the role says.
func TestGatewayClaimsCanPublishFromScope(t *testing.T) {
	cases := map[string]struct {
		scope string
		want  bool
	}{
		"both grants":                  {scope: "p2p:publish p2p:subscribe", want: true},
		"publish alone":                {scope: "p2p:publish", want: true},
		"subscribe only is read-only":  {scope: "p2p:subscribe", want: false},
		"unknown grant alone":          {scope: "p2p:admin", want: false},
		"unknown grant beside publish": {scope: "p2p:admin p2p:publish", want: true},
		"order does not matter":        {scope: "p2p:subscribe p2p:publish", want: true},
		"surrounding spaces":           {scope: "  p2p:subscribe p2p:publish  ", want: true},
		"case does not fold":           {scope: "P2P:PUBLISH", want: false},
		"subscribe constant alone":     {scope: entities.GrantP2PSubscribe, want: false},
		// Only SP separates, per RFC 6749 §3.3. Anything else leaves one unknown token,
		// which is what stops a grant being smuggled through an exotic space.
		"commas do not separate":   {scope: "p2p:subscribe,p2p:publish", want: false},
		"tabs do not separate":     {scope: "p2p:subscribe\tp2p:publish", want: false},
		"NBSP does not separate":   {scope: "p2p:subscribe\u00a0p2p:publish", want: false},
		"U+2028 does not separate": {scope: "p2p:subscribe\u2028p2p:publish", want: false},
		// Nor at the edges: exotic space is part of the token, not trimmed off it.
		"leading NBSP sticks":           {scope: "\u00a0p2p:publish", want: false},
		"trailing newline sticks":       {scope: "p2p:publish\n", want: false},
		"lone tab is one unknown grant": {scope: "\t", want: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// hermes so the role would publish: a stated grant has to override it.
			claims := entities.GatewayClaims{Type: entities.GatewayTypeHermes, Scope: tc.scope}
			require.Equal(t, tc.want, claims.CanPublish())
		})
	}
}

// The fallback, for tokens minted before the claim existed. Deleted once every
// credential carries scope, at which point no real token changes behavior.
func TestGatewayClaimsCanPublishFallsBackToRole(t *testing.T) {
	cases := map[string]struct {
		gatewayType entities.GatewayType
		want        bool
	}{
		"hermes":       {gatewayType: entities.GatewayTypeHermes, want: true},
		"partner":      {gatewayType: entities.GatewayTypePartner, want: true},
		"relay":        {gatewayType: entities.GatewayTypeRelay, want: true},
		"stream":       {gatewayType: "stream", want: false},
		"unknown role": {gatewayType: "future-role", want: false},
		"empty role":   {gatewayType: "", want: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Only SP collapses to nothing. Exotic whitespace is a token, tested above.
			for _, scope := range []string{"", "   "} {
				claims := entities.GatewayClaims{Type: tc.gatewayType, Scope: scope}
				require.Equal(t, tc.want, claims.CanPublish(), "scope %q", scope)
			}
		})
	}
}

// Whole tokens, never substrings or prefixes. `p2p:publish:<topic>` is the extension
// ADR-0001 contemplates: it must be minted alongside `p2p:publish`, not instead of it.
func TestGatewayClaimsCanPublishMatchesWholeTokens(t *testing.T) {
	for _, scope := range []string{"p2p:publish:beacon_block", "p2p:publish:", "p2p:pub", "xp2p:publish"} {
		claims := entities.GatewayClaims{Type: entities.GatewayTypeHermes, Scope: scope}
		require.False(t, claims.CanPublish(), "scope %q", scope)
	}
}

// The other direction of precedence: grants decide, so a role that cannot publish is
// irrelevant once the claim states one. The table above covers the reverse.
func TestGatewayClaimsGrantOverridesNonPublishingRole(t *testing.T) {
	claims := entities.GatewayClaims{Type: "stream", Scope: entities.GrantP2PPublish}
	require.True(t, claims.CanPublish())
}

// A grant from some future unrelated vocabulary still disables the fallback, so it
// must not reach this token until every credential carries its p2p grants.
func TestGatewayClaimsUnrelatedGrantDisablesTheFallback(t *testing.T) {
	partner := entities.GatewayClaims{Type: entities.GatewayTypePartner, Scope: "obs:read"}
	require.False(t, partner.CanPublish())
}

// Only `type` and `scope` decide, which is what lets the doc put aud on the caller.
func TestGatewayClaimsCanPublishIgnoresOtherClaims(t *testing.T) {
	body := `{"aud":["p2p","services","stream"],"type":"stream","scope":"p2p:subscribe","scope_version":3,` +
		`"chain_id":"hoodi","operator_id":"op-1","cnf":{"peer_id":"12D3Koo"}}`
	var claims entities.GatewayClaims
	require.NoError(t, json.Unmarshal([]byte(body), &claims))
	require.False(t, claims.CanPublish())

	granted := claims
	granted.Scope = entities.GrantP2PPublish
	require.True(t, granted.CanPublish())
}

func TestGatewayClaimsCanPublishNilReceiver(t *testing.T) {
	var claims *entities.GatewayClaims
	require.False(t, claims.CanPublish())
}

// The fallback is authorization-relevant, so the role constants are pinned against the
// wire rather than only against themselves. Decoding does not fold case or trim.
func TestGatewayClaimsFallbackReadsRoleFromJSON(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{body: `{"type":"hermes"}`, want: true},
		{body: `{"type":"partner"}`, want: true},
		{body: `{"type":"relay"}`, want: true},
		{body: `{"type":"stream"}`, want: false},
		{body: `{"type":"Hermes"}`, want: false},
		{body: `{"type":" hermes "}`, want: false},
	} {
		var claims entities.GatewayClaims
		require.NoError(t, json.Unmarshal([]byte(tc.body), &claims))
		require.Equal(t, tc.want, claims.CanPublish(), tc.body)
	}
}

// scope is a STRING per RFC 8693 §4.2. A []string field would fail to decode the real
// claim, so the wire type is pinned here rather than discovered against optimum-auth.
func TestGatewayClaimsScopeUnmarshalsFromString(t *testing.T) {
	var claims entities.GatewayClaims
	require.NoError(t, json.Unmarshal([]byte(`{"scope":"p2p:publish p2p:subscribe"}`), &claims))
	require.Equal(t, "p2p:publish p2p:subscribe", claims.Scope)
	require.True(t, claims.CanPublish())

	require.Error(t, json.Unmarshal([]byte(`{"scope":["p2p:publish"]}`), &entities.GatewayClaims{}))
}

// A string field cannot tell an omitted claim from "", and the issuer relies on that:
// it omits instead of emitting empty, so both reach the role fallback.
func TestGatewayClaimsExplicitlyEmptyScopeReadsAsAbsent(t *testing.T) {
	var present, omitted, null entities.GatewayClaims
	require.NoError(t, json.Unmarshal([]byte(`{"type":"hermes","scope":""}`), &present))
	require.NoError(t, json.Unmarshal([]byte(`{"type":"hermes"}`), &omitted))
	require.NoError(t, json.Unmarshal([]byte(`{"type":"hermes","scope":null}`), &null))
	require.Equal(t, omitted.CanPublish(), present.CanPublish())
	require.Equal(t, omitted.CanPublish(), null.CanPublish())
	require.True(t, present.CanPublish())
}

// omitempty, so an audience with no grants to state does not put an empty `scope` on
// the wire (as for cnf and operator_id).
func TestGatewayClaimsScopeOmittedWhenEmpty(t *testing.T) {
	out, err := json.Marshal(entities.GatewayClaims{})
	require.NoError(t, err)
	require.NotContains(t, string(out), `"scope"`)
}
