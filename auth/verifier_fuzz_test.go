package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

const testKID = "k1"

func FuzzVerifierVerify(f *testing.F) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	// wrap the public key into jwks
	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
		testKID: keyfunc.NewGivenRSA(&key.PublicKey, keyfunc.GivenKeyOptions{}),
	})
	v := &Verifier{jwks: jwks}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "seed"})
	tok.Header["kid"] = testKID
	signed, _ := tok.SignedString(key)
	// seed inputs
	f.Add(signed)
	f.Add("invalid.string")
	// fuzzing function
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = v.Verify(s)
	})
}

func FuzzNewVerifierFromDomain(f *testing.F) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")
	f.Add(host)
	f.Fuzz(func(t *testing.T, domain string) {
		if domain != host {
			t.Skip()
		}
		_, _ = NewVerifierFromDomain(domain, "", nil)
	})
}
