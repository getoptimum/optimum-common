package entities

import "github.com/golang-jwt/jwt/v5"

// TokenAudience is a gateway-JWT `aud` value. optimum-auth mints one token
// per audience off the same key: p2p (handshake), services (bootstrap),
// stream (ADR-0011 consumers).
type TokenAudience string

const (
	TokenAudienceP2P      TokenAudience = "p2p"
	TokenAudienceServices TokenAudience = "services"
	TokenAudienceStream   TokenAudience = "stream"
)

func (a TokenAudience) String() string {
	return string(a)
}

// GatewayConfirmation is the RFC 7800 `cnf` claim binding the JWT to the
// gateway's libp2p peer_id, so a replayed bearer token can be rejected.
type GatewayConfirmation struct {
	PeerID string `json:"peer_id"`
}

// GatewayClaims is the superset of gateway-JWT claims across all audiences;
// each token carries only the subset that applies to it.
type GatewayClaims struct {
	ScopeVersion int64       `json:"scope_version"`
	Type         GatewayType `json:"type"`
	ChainID      string      `json:"chain_id,omitempty"`
	// Capability grants on the peer-visible p2p token: RFC 8693 §4.2, space-separated,
	// never an array. Omitted and "" alike mean a pre-scope mint; see CanPublish.
	Scope string `json:"scope,omitempty"`
	// Set only on the services token; must never leak onto the peer-visible
	// p2p handshake token (optimum-bootstrap#262).
	OperatorID string `json:"operator_id,omitempty"`
	// Pointer so omitempty drops it entirely; a value struct would still emit
	// "cnf":{"peer_id":""} on cnf-less handshake tokens.
	CNF *GatewayConfirmation `json:"cnf,omitempty"`
	jwt.RegisteredClaims
}

// HasAudience reports whether the token's `aud` contains want. aud is
// multi-valued per RFC 7519, so membership is the spec-correct check.
func (c *GatewayClaims) HasAudience(want TokenAudience) bool {
	for _, a := range c.Audience {
		if a == want.String() {
			return true
		}
	}
	return false
}
