package auth_test

import (
	"testing"

	ocauth "github.com/getoptimum/optimum-common/auth"
	"github.com/golang-jwt/jwt/v5"
)

func FuzzParseUnverified(f *testing.F) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "seed"}).SignedString([]byte("secret"))
	f.Add(token)
	f.Add("not.a.jwt")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ocauth.ParseUnverified(s)
	})
}
