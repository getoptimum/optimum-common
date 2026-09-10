package entities

import (
	"fmt"
	"strings"
)

// GatewayType is the per-key role. Publish/subscribe is GatewayClaims.Scope;
// Type.CanPublish is only the fallback for tokens minted before scope existed.
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

// CanPublish is the pre-scope fallback. Handshake must use GatewayClaims.CanPublish.
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
