package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// FuzzParseUnverified ensures the parser handles arbitrary input without panicking.
func FuzzParseUnverified(f *testing.F) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "seed"}).SignedString([]byte("secret"))
	f.Add(token)
	f.Add("not.a.jwt")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ParseUnverified(s)
	})
}
