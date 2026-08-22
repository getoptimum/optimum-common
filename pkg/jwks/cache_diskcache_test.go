package jwks_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	commonio "github.com/getoptimum/optimum-common/pkg/io"
	"github.com/getoptimum/optimum-common/pkg/jwks"
	"github.com/getoptimum/optimum-common/pkg/logger"
)

// switchableJWKS is a JWKS endpoint whose body can be swapped mid-test, so a
// healthy provider can start returning something that is not a JWK set.
type switchableJWKS struct {
	server *httptest.Server
	body   atomic.Pointer[[]byte]
	served atomic.Int32
}

func newSwitchableJWKS(t *testing.T) (*switchableJWKS, []byte) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	good, err := json.Marshal(map[string]any{
		"keys": []map[string]string{{
			"kty": "EC", "crv": "P-256", "kid": "test-key",
			"alg": "ES256", "use": "sig",
			"x": base64.RawURLEncoding.EncodeToString(priv.X.Bytes()),
			"y": base64.RawURLEncoding.EncodeToString(priv.Y.Bytes()),
		}},
	})
	require.NoError(t, err)

	s := &switchableJWKS{}
	s.body.Store(&good)
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.served.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(*s.body.Load())
	}))
	t.Cleanup(s.server.Close)

	return s, good
}

func TestRefresh_KeepsDiskCacheBootableWhenWireGoesBad(t *testing.T) {
	// Given a healthy provider that has seeded the disk cache.
	s, good := newSwitchableJWKS(t)
	diskPath := filepath.Join(t.TempDir(), "jwks.json")

	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL:  s.server.URL,
		DiskPath: diskPath,
		Refresh:  20 * time.Millisecond,
	})
	require.NoError(t, err)

	persisted, err := commonio.LoadFromFile(diskPath)
	require.NoError(t, err)
	require.JSONEq(t, string(good), string(persisted))

	// When the provider starts serving a key set whose key does not decode.
	// (A document like {"not":"a jwk set"} parses fine, so it would not
	// exercise this path — the payload has to actually fail to parse.)
	garbage := []byte(`{"keys":[{"kty":"EC","crv":"P-256","kid":"test-key","alg":"ES256","use":"sig","x":"!!!not-base64!!!","y":"!!!not-base64!!!"}]}`)
	s.body.Store(&garbage)
	after := s.served.Load()

	require.Eventually(t, func() bool {
		return s.served.Load() >= after+2
	}, 2*time.Second, 20*time.Millisecond, "refresh loop did not run")

	// Then the disk cache still holds the last payload that actually parsed,
	// so the next cold boot still comes up.
	persisted, err = commonio.LoadFromFile(diskPath)
	require.NoError(t, err)
	require.JSONEq(t, string(good), string(persisted))
}

func TestNew_FallsBackToDisk_WhenWireServesUnparseableJWKS(t *testing.T) {
	// Given a disk cache seeded while the provider was healthy.
	s, good := newSwitchableJWKS(t)
	diskPath := filepath.Join(t.TempDir(), "jwks.json")

	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: s.server.URL, DiskPath: diskPath, Refresh: time.Hour,
	})
	require.NoError(t, err)

	// When the provider is reachable at boot but answers with a key that
	// does not decode.
	garbage := []byte(`{"keys":[{"kty":"EC","crv":"P-256","kid":"test-key","alg":"ES256","use":"sig","x":"!!!not-base64!!!","y":"!!!not-base64!!!"}]}`)
	s.body.Store(&garbage)

	c, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: s.server.URL, DiskPath: diskPath, Refresh: time.Hour,
	})

	// Then boot succeeds from disk rather than failing outright.
	require.NoError(t, err)
	require.NotNil(t, c)

	persisted, readErr := commonio.LoadFromFile(diskPath)
	require.NoError(t, readErr)
	require.JSONEq(t, string(good), string(persisted))
}
