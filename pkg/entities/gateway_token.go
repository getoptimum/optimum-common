package entities

import "github.com/golang-jwt/jwt/v5"

// TokenAudience is the `aud` claim on a gateway JWT. optimum-auth mints two
// tokens with the same signing key and issuer but different audiences so each
// recipient can require the one intended for it (RFC 7519 §4.1.3):
//   - TokenAudienceP2P: the peer-visible handshake token, presented at the
//     libp2p stream handshake. Minimal claims only (privacy boundary).
//   - TokenAudienceServices: the services token, which carries operator_id and
//     authenticates to centralized services (bootstrap ingest, the Mimir/Loki
//     push proxy). Never presented to peers.
//
// Defining the values here lets the gateway, bootstrap, and billing verifiers
// compare against one source instead of re-declaring the strings.
type TokenAudience string

const (
	TokenAudienceP2P      TokenAudience = "p2p"
	TokenAudienceServices TokenAudience = "services"
)

func (a TokenAudience) String() string {
	return string(a)
}

// GatewayConfirmation is the RFC 7800 `cnf` claim, populated with the gateway's
// libp2p peer_id at mint. Verifiers compare cnf.peer_id to the authenticated
// libp2p remote peer to bind the JWT to one node and block bearer-token replay.
type GatewayConfirmation struct {
	PeerID string `json:"peer_id"`
}

// GatewayClaims is the claim set on a gateway JWT minted by optimum-auth. It is
// the superset across the two audiences; a given token carries the subset that
// applies to it:
//   - scope_version / type / chain_id: present on both tokens.
//   - operator_id: present ONLY on the services token (aud=services). It
//     identifies the operator and MUST never appear on the peer-visible
//     handshake token (aud=p2p). See optimum-bootstrap#262.
//   - cnf: present on both when a peer_id was supplied at mint.
//
// The `aud` value is the embedded RegisteredClaims.Audience; compare it against
// the TokenAudience constants rather than a bare string literal.
type GatewayClaims struct {
	ScopeVersion int64               `json:"scope_version"`
	Type         GatewayType         `json:"type"`
	ChainID      string              `json:"chain_id,omitempty"`
	OperatorID   string              `json:"operator_id,omitempty"`
	CNF          GatewayConfirmation `json:"cnf,omitempty"`
	jwt.RegisteredClaims
}

// HasAudience reports whether the token's `aud` claim contains want. A gateway
// JWT carries a single audience, but `aud` is multi-valued per RFC 7519, so the
// membership check is the spec-correct comparison.
func (c *GatewayClaims) HasAudience(want TokenAudience) bool {
	for _, a := range c.Audience {
		if a == want.String() {
			return true
		}
	}
	return false
}
