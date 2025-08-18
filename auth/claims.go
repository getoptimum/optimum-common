package auth

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt"
)

// Claims is the single shared model used everywhere.
// As of now: for cli and proxy
type Claims struct {
	Subject           string    `json:"sub"`                  // subject
	IssuedAt          time.Time `json:"iat"`                  // issued at
	ExpiresAt         time.Time `json:"exp"`                  // expiry
	IsActive          bool      `json:"is_active"`            // active flag
	MaxPublishPerHour int       `json:"max_publish_per_hour"` // per-hour cap
	MaxPublishPerSec  int       `json:"max_publish_per_sec"`  // per-sec cap
	MaxMessageSize    int64     `json:"max_message_size"`     // bytes
	DailyQuota        int64     `json:"daily_quota"`          // messages/day
	ClientID          string    `json:"client_id"`            // client id
	LimitsSetAt       int64     `json:"limits_set_at"`        // unix seconds
	Tier              string    `json:"tier"`                 // plan tier
}

// LimitsDefaults are optional fallbacks applied when a claim is absent.
type LimitsDefaults struct {
	MaxPublishPerHour int
	MaxPublishPerSec  int
	MaxMessageSize    int64
	DailyQuota        int64
}

var defaultLimits LimitsDefaults

// SetDefaultLimits globally sets fallbacks (optional).
// If you don't call this, omitted limit claims remain zero.
func SetDefaultLimits(d LimitsDefaults) { defaultLimits = d }

// fromMap builds Claims from jwt.MapClaims and applies defaults.
// Detailed desc:
// Converts jwt.MapClaims into a Claims struct.
// Handles missing fields by using defaults from SetDefaultLimits.
// Falls back to Subject for ClientID if client_id is missing.
// Parses iat and exp into time.Time
func fromMap(mc jwt.MapClaims) (*Claims, error) {
	c := &Claims{
		Subject:           str(mc, "sub", ""),
		IsActive:          boolv(mc, "is_active", false),
		MaxPublishPerHour: intv(mc, "max_publish_per_hour", defaultLimits.MaxPublishPerHour),
		MaxPublishPerSec:  intv(mc, "max_publish_per_sec", defaultLimits.MaxPublishPerSec),
		MaxMessageSize:    int64v(mc, "max_message_size", defaultLimits.MaxMessageSize),
		DailyQuota:        int64v(mc, "daily_quota", defaultLimits.DailyQuota),
		ClientID:          str(mc, "client_id", ""),
		LimitsSetAt:       int64v(mc, "limits_set_at", 0),
		Tier:              str(mc, "tier", ""),
	}
	if c.ClientID == "" && c.Subject != "" {
		c.ClientID = c.Subject // fallback
	}
	// times (iat/exp) are NumericDate in JWT — commonly float
	if iat, ok := mc["iat"]; ok {
		if sec, ok := toInt64(iat); ok {
			c.IssuedAt = time.Unix(sec, 0)
		}
	}
	if exp, ok := mc["exp"]; ok {
		if sec, ok := toInt64(exp); ok {
			c.ExpiresAt = time.Unix(sec, 0)
		}
	}
	return c, nil
}

// helpers tolerant to float64, json.Number, string
func toInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case string:
		n := json.Number(t)
		if i, err := n.Int64(); err == nil {
			return i, true
		}
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return i, true
		}
	}
	return 0, false
}

// Note: in case many more field types added -> use generics
func int64v(mc jwt.MapClaims, key string, def int64) int64 {
	if v, ok := mc[key]; ok {
		if i, ok := toInt64(v); ok {
			return i
		}
	}
	return def
}

func intv(mc jwt.MapClaims, key string, def int) int {
	if v, ok := mc[key]; ok {
		if i, ok := toInt64(v); ok {
			return int(i)
		}
	}
	return def
}

func str(mc jwt.MapClaims, key, def string) string {
	if v, ok := mc[key].(string); ok {
		return v
	}
	return def
}

func boolv(mc jwt.MapClaims, key string, def bool) bool {
	if v, ok := mc[key].(bool); ok {
		return v
	}
	return def
}
