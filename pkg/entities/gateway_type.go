package entities

import (
	"fmt"
	"strings"
)

// GatewayType is the per-key gateway role. The role fully determines
// the publish/subscribe matrix on mump2p (see ADR-004 §Gateway types) — billing
// does not mint per-key topic lists; verifiers hard-code the mapping.
type GatewayType string

const (
	GatewayTypeHermes  GatewayType = "hermes"
	GatewayTypePartner GatewayType = "partner"
	GatewayTypeRelay   GatewayType = "relay"
	GatewayTypeStream  GatewayType = "stream"
)

var (
	gatewayTypeMapper = map[string]GatewayType{
		"hermes":  GatewayTypeHermes,
		"partner": GatewayTypePartner,
		"relay":   GatewayTypeRelay,
		"stream":  GatewayTypeStream,
	}
)

func (s GatewayType) String() string {
	return string(s)
}

// CanPublish is whether this role may originate mump2p traffic. Fail-closed:
// only hermes/partner/relay publish; stream, empty, and unknown do not.
func (s GatewayType) CanPublish() bool {
	switch s {
	case GatewayTypeHermes, GatewayTypePartner, GatewayTypeRelay:
		return true
	default:
		return false
	}
}

func GatewayTypeFromString(s string) (GatewayType, error) {
	if val, ok := gatewayTypeMapper[strings.TrimSpace(strings.ToLower(s))]; ok {
		return val, nil
	}
	return "", fmt.Errorf("unknown gateway type: %q", s)
}
