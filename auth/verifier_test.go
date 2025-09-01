package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestVerifier_VerifyFailures exercises audience/issuer mismatches and time-based
// validations including leeway handling.
func TestVerifier_VerifyFailures(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
		"k1": keyfunc.NewGivenRSA(&key.PublicKey, keyfunc.GivenKeyOptions{}),
	})
	v := &Verifier{jwks: jwks, Audience: "aud", Issuer: "iss"}

	baseClaims := jwt.MapClaims{
		"sub": "alice",
		"aud": "aud",
		"iss": "iss",
		"exp": time.Now().Add(time.Minute).Unix(),
	}

	sign := func(claims jwt.MapClaims) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tok.Header["kid"] = "k1"
		s, err := tok.SignedString(key)
		require.NoError(t, err)
		return s
	}

	// audience mismatch
	audClaims := maps.Clone(baseClaims)
	audClaims["aud"] = "other"
	_, err = v.Verify(sign(audClaims))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidAudience)

	// issuer mismatch
	issClaims := maps.Clone(baseClaims)
	issClaims["iss"] = "other"
	_, err = v.Verify(sign(issClaims))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidIssuer)

	// expired beyond leeway
	expClaims := maps.Clone(baseClaims)
	expClaims["exp"] = time.Now().Add(-2 * time.Minute).Unix()
	_, err = v.Verify(sign(expClaims))
	require.Error(t, err)

	// not yet valid but within leeway should pass
	nbfClaims := maps.Clone(baseClaims)
	nbfClaims["nbf"] = time.Now().Add(20 * time.Second).Unix()
	_, err = v.Verify(sign(nbfClaims))
	require.NoError(t, err)

	// not yet valid beyond leeway should fail
	nbfClaims["nbf"] = time.Now().Add(time.Minute).Unix()
	_, err = v.Verify(sign(nbfClaims))
	require.Error(t, err)
}

// TestNewVerifierFromDomain_FetchError ensures fetch failures are surfaced and
// JWKS refresh errors invoke the handler.
func TestNewVerifierFromDomain_FetchAndRefresh(t *testing.T) {
	// JWKS for test server
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())
	jwksJSON := []byte(`{"keys":[{"kty":"RSA","alg":"RS256","use":"sig","kid":"k1","n":"` + n + `","e":"` + e + `"}]}`)

	var calls int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jwksJSON)
			return
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "https://")
	errCh := make(chan error, 1)
	v, err := NewVerifierFromDomain(host, "aud", &VerifierOptions{
		RefreshInterval:     50 * time.Millisecond,
		HTTPClient:          srv.Client(),
		RefreshErrorHandler: func(err error) { errCh <- err },
	})
	require.NoError(t, err)
	require.NotNil(t, v)

	select {
	case <-errCh:
		// expected refresh error
	case <-time.After(2 * time.Second):
		t.Fatal("expected refresh error")
	}

	// ensure fetch error from invalid domain is returned
	_, err = NewVerifierFromDomain("invalid.domain", "aud", nil)
	require.Error(t, err)
}
