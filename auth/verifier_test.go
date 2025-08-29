package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// TestVerifier_Verify ensures tokens signed with keys in the JWKS are validated.
func TestVerifier_Verify(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
		"k1": keyfunc.NewGivenRSA(&key.PublicKey, keyfunc.GivenKeyOptions{}),
	})
	v := &Verifier{jwks: jwks, Audience: "aud", Issuer: "iss"}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "alice",
		"aud": "aud",
		"iss": "iss",
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	token.Header["kid"] = "k1"
	signed, err := token.SignedString(key)
	require.NoError(t, err)

	claims, err := v.Verify(signed)
	require.NoError(t, err)
	require.Equal(t, "alice", claims.Subject)
}
