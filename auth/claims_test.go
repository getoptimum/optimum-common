package auth_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestFromMap_DefaultsAndFallback(t *testing.T) {
	auth.SetDefaultLimits(auth.LimitsDefaults{MaxPublishPerHour: 10, MaxPublishPerSec: 5, MaxMessageSize: 2048, DailyQuota: 1000})
	defer auth.SetDefaultLimits(auth.LimitsDefaults{})

	mc := jwt.MapClaims{
		"sub": "alice",
		"iat": float64(100),
		"exp": float64(200),
	}
	c, err := auth.FromMap(mc)
	require.NoError(t, err)

	require.Equal(t, "alice", c.Subject)
	// if clientID not present  -> defaults to SubjectID
	require.Equal(t, "alice", c.ClientID)
	require.Equal(t, time.Unix(100, 0), c.IssuedAt)
	require.Equal(t, time.Unix(200, 0), c.ExpiresAt)
	require.Equal(t, 10, c.MaxPublishPerHour)
	require.Equal(t, 5, c.MaxPublishPerSec)
	require.Equal(t, int64(2048), c.MaxMessageSize)
	require.Equal(t, int64(1000), c.DailyQuota)
}

func TestFromMap_Coercion(t *testing.T) {
	mc := jwt.MapClaims{
		"sub":                  "bob",
		"client_id":            "client",
		"is_active":            true,
		"max_publish_per_hour": float64(1),
		"max_publish_per_sec":  json.Number("2"),
		"max_message_size":     "3",
		"daily_quota":          json.Number("4"),
		"limits_set_at":        "5",
		"iat":                  json.Number("1000"),
		"exp":                  "2000",
		"tier":                 "gold",
	}

	c, err := auth.FromMap(mc)
	require.NoError(t, err)

	require.Equal(t, "bob", c.Subject)
	require.Equal(t, "client", c.ClientID)
	require.True(t, c.IsActive)
	require.Equal(t, 1, c.MaxPublishPerHour)
	require.Equal(t, 2, c.MaxPublishPerSec)
	require.Equal(t, int64(3), c.MaxMessageSize)
	require.Equal(t, int64(4), c.DailyQuota)
	require.Equal(t, int64(5), c.LimitsSetAt)
	require.Equal(t, time.Unix(1000, 0), c.IssuedAt)
	require.Equal(t, time.Unix(2000, 0), c.ExpiresAt)
	require.Equal(t, "gold", c.Tier)
}

func TestParseUnverified(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user"}).SignedString([]byte("secret"))
	require.NoError(t, err)

	c, err := auth.ParseUnverified(token)
	require.NoError(t, err)
	require.Equal(t, "user", c.Subject)
}

func TestParseUnverified_Invalid(t *testing.T) {
	_, err := auth.ParseUnverified("not.a.jwt")
	require.Error(t, err)
}
