package entities

import (
	"slices"
	"strings"
)

// Capability grants carried in the `scope` claim (optimum-auth ADR-0001).
const (
	GrantP2PPublish   = "p2p:publish"
	GrantP2PSubscribe = "p2p:subscribe"
)

// grants splits the claim on SP alone, per RFC 6749 §3.3. A wider separator class would
// cut `p2p:subscribe<NBSP>p2p:publish` into two grants instead of one unrecognized one.
func (c *GatewayClaims) grants() []string {
	return strings.FieldsFunc(c.Scope, func(r rune) bool { return r == ' ' })
}

// CanPublish reports whether the peer may originate mump2p traffic. Grants decide it;
// the role is the fallback for a pre-scope token. Verify the signature and aud first.
func (c *GatewayClaims) CanPublish() bool {
	if c == nil {
		return false
	}
	grants := c.grants()
	if len(grants) == 0 {
		// Any grant at all disables this fallback, so none unrelated to p2p may be
		// minted onto this token until the p2p grants have rolled out everywhere.
		return c.Type.CanPublish()
	}
	return slices.Contains(grants, GrantP2PPublish)
}
