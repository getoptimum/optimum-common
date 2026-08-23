package jwks_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	commonio "github.com/getoptimum/optimum-common/pkg/io"
	"github.com/getoptimum/optimum-common/pkg/jwks"
	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type jwksTestRig struct {
	priv   *ecdsa.PrivateKey
	server *httptest.Server
	calls  atomic.Int32
}

func newRig(t *testing.T) *jwksTestRig {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	rig := &jwksTestRig{priv: priv}
	jwksDoc, err := json.Marshal(map[string]any{
		"keys": []map[string]string{{
			"kty": "EC", "crv": "P-256", "kid": "test-key",
			"alg": "ES256", "use": "sig",
			"x": base64.RawURLEncoding.EncodeToString(priv.X.Bytes()),
			"y": base64.RawURLEncoding.EncodeToString(priv.Y.Bytes()),
		}},
	})
	require.NoError(t, err)

	rig.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		rig.calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksDoc)
	}))
	t.Cleanup(rig.server.Close)
	return rig
}

func (r *jwksTestRig) sign(t *testing.T) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": "test-subject",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = "test-key"
	signed, err := tok.SignedString(r.priv)
	require.NoError(t, err)
	return signed
}

func TestNew_RequiresLogger(t *testing.T) {
	_, err := jwks.New(t.Context(), nil, jwks.Config{
		JWKSURL: "https://example.com/jwks.json", DiskPath: "/tmp/x.json", Refresh: time.Hour,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "log")
}

func TestNew_RequiresJWKSURL(t *testing.T) {
	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		DiskPath: "/tmp/x.json", Refresh: time.Hour,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWKSURL")
}

func TestNew_RequiresDiskPath(t *testing.T) {
	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: "https://example.com/jwks.json", Refresh: time.Hour,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "DiskPath")
}

func TestNew_RequiresRefresh(t *testing.T) {
	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: "https://example.com/jwks.json", DiskPath: "/tmp/x.json",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Refresh")
}

func TestNew_SeedsFromWire_AndKeyfuncWorks(t *testing.T) {
	rig := newRig(t)
	diskPath := filepath.Join(t.TempDir(), "jwks.json")

	c, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL:  rig.server.URL,
		DiskPath: diskPath,
		Refresh:  time.Hour,
	})
	require.NoError(t, err)

	token, _, err := new(jwt.Parser).ParseUnverified(rig.sign(t), jwt.MapClaims{})
	require.NoError(t, err)
	key, err := c.Keyfunc(token)
	require.NoError(t, err)
	require.NotNil(t, key)

	// And the persisted disk copy exists, so a future cold boot survives a
	// wire outage.
	_, err = os.Stat(diskPath)
	require.NoError(t, err)
}

func TestNew_FallsBackToDisk_WhenWireDown(t *testing.T) {
	rig := newRig(t)
	diskPath := filepath.Join(t.TempDir(), "jwks.json")

	// First boot: wire is up; populate the disk cache.
	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: rig.server.URL, DiskPath: diskPath, Refresh: time.Hour,
	})
	require.NoError(t, err)
	rig.server.Close()

	// Second boot: wire is down (server closed), but the disk cache exists.
	c, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: rig.server.URL, DiskPath: diskPath, Refresh: time.Hour,
	})
	require.NoError(t, err, "disk fallback must let us boot when wire is down")

	token, _, err := new(jwt.Parser).ParseUnverified(rig.sign(t), jwt.MapClaims{})
	require.NoError(t, err)
	key, err := c.Keyfunc(token)
	require.NoError(t, err)
	require.NotNil(t, key)
}

func TestNew_FailsWhenWireDown_AndNoDisk(t *testing.T) {
	rig := newRig(t)
	rig.server.Close() // wire down before construction

	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL:  rig.server.URL,
		DiskPath: filepath.Join(t.TempDir(), "missing.json"),
		Refresh:  time.Hour,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "disk cache unavailable")
}

// switchableJWKS serves a JWKS document that can be swapped mid-test, so a
// healthy provider can start returning a key set that does not decode.
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

// A document like {"not":"a jwk set"} parses fine, so the payload has to be a
// key set whose key does not decode for this path to be exercised at all.
const undecodableJWKS = `{"keys":[{"kty":"EC","crv":"P-256","kid":"test-key","alg":"ES256","use":"sig","x":"!!!not-base64!!!","y":"!!!not-base64!!!"}]}`

func TestRefresh_KeepsDiskCacheBootableWhenWireGoesBad(t *testing.T) {
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

	garbage := []byte(undecodableJWKS)
	s.body.Store(&garbage)
	after := s.served.Load()

	require.Eventually(t, func() bool {
		return s.served.Load() >= after+2
	}, 2*time.Second, 20*time.Millisecond, "refresh loop did not run")

	// The disk copy still holds the last payload that parsed, so the next cold
	// boot still comes up.
	persisted, err = commonio.LoadFromFile(diskPath)
	require.NoError(t, err)
	require.JSONEq(t, string(good), string(persisted))
}

func TestNew_FallsBackToDisk_WhenWireServesUnparseableJWKS(t *testing.T) {
	s, good := newSwitchableJWKS(t)
	diskPath := filepath.Join(t.TempDir(), "jwks.json")

	_, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: s.server.URL, DiskPath: diskPath, Refresh: time.Hour,
	})
	require.NoError(t, err)

	garbage := []byte(undecodableJWKS)
	s.body.Store(&garbage)

	c, err := jwks.New(t.Context(), logger.NewAppSLogger(logger.Debug), jwks.Config{
		JWKSURL: s.server.URL, DiskPath: diskPath, Refresh: time.Hour,
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	persisted, readErr := commonio.LoadFromFile(diskPath)
	require.NoError(t, readErr)
	require.JSONEq(t, string(good), string(persisted))
}
