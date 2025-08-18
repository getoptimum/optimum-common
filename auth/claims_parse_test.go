// Primary functions to cover
// fromMap()
// ParseUnverified()
// Helper functions
package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func fakeToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("fake"))
	return signed
}

func TestParseUnverified(t *testing.T) {
	now := time.Now().Unix()

	// Set shared defaults for tests
	SetDefaultLimits(LimitsDefaults{
		MaxPublishPerHour: 111,
		MaxPublishPerSec:  7,
		MaxMessageSize:    2048,
		DailyQuota:        4096,
	})

	tests := []struct {
		name      string
		claims    jwt.MapClaims
		expectErr bool
		expect    Claims
	}{
		{
			name: "all fields present",
			claims: jwt.MapClaims{
				"sub":                  "user123",
				"iat":                  float64(now),
				"exp":                  float64(now + 3600),
				"client_id":            "cli-app",
				"limits_set_at":        float64(now - 100),
				"is_active":            true,
				"max_publish_per_hour": float64(100),
				"max_publish_per_sec":  float64(2),
				"max_message_size":     float64(1048576),
				"daily_quota":          float64(1073741824),
				"tier":                 "pro",
			},
			expectErr: false,
			expect: Claims{
				Subject:           "user123",
				ClientID:          "cli-app",
				IsActive:          true,
				MaxPublishPerHour: 100,
				MaxPublishPerSec:  2,
				MaxMessageSize:    1048576,
				DailyQuota:        1073741824,
				Tier:              "pro",
			},
		},
		{
			name: "missing optional fields",
			claims: jwt.MapClaims{
				"sub":       "anon",
				"is_active": false,
			},
			expectErr: false,
			expect: Claims{
				Subject:           "anon",
				ClientID:          "anon", // fallback from sub
				IsActive:          false,
				MaxPublishPerHour: 111,
				MaxPublishPerSec:  7,
				MaxMessageSize:    2048,
				DailyQuota:        4096,
			},
		},
		{
			name:      "malformed token",
			claims:    nil,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tokenStr string
			if tc.claims != nil {
				tokenStr = fakeToken(tc.claims)
			} else {
				tokenStr = "not.a.valid.token"
			}

			claims, err := ParseUnverified(tokenStr)

			if tc.expectErr {
				require.Error(t, err)
				require.Nil(t, claims)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)

			require.Equal(t, tc.expect.Subject, claims.Subject)
			require.Equal(t, tc.expect.ClientID, claims.ClientID)
			require.Equal(t, tc.expect.IsActive, claims.IsActive)
			require.Equal(t, tc.expect.MaxPublishPerHour, claims.MaxPublishPerHour)
			require.Equal(t, tc.expect.MaxPublishPerSec, claims.MaxPublishPerSec)
			require.Equal(t, tc.expect.MaxMessageSize, claims.MaxMessageSize)
			require.Equal(t, tc.expect.DailyQuota, claims.DailyQuota)
			require.Equal(t, tc.expect.Tier, claims.Tier)
		})
	}
}
