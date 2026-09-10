package entities

import (
	"fmt"
	"strings"
)

// GatewayType is the per-key gateway commercial role. It no longer decides what a peer
// may do on mump2p (`scope` does, optimum-auth ADR-0001) except as the fallback below.
type GatewayType string

const (
	GatewayTypeHermes  GatewayType = "hermes"
	GatewayTypePartner GatewayType = "partner"
	GatewayTypeRelay   GatewayType = "relay"
)

var gatewayTypeMapper = map[string]GatewayType{
	"hermes":  GatewayTypeHermes,
	"partner": GatewayTypePartner,
	"relay":   GatewayTypeRelay,
}

func (s GatewayType) String() string {
	return string(s)
}

// CanPublish is the pre-scope fallback; callers must migrate to GatewayClaims.CanPublish.
// Fail-closed: an unrecognized or empty role publishes nothing.
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
