package netutil_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/netutil"
	"github.com/stretchr/testify/require"
)

func TestGetOutboundP2PAddr(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ip":"5.6.7.8"}`))
	}))
	defer srv.Close()
	t.Setenv("REMOTE_HOST", strings.TrimPrefix(srv.URL, "http://"))

	addr, err := netutil.GetOutboundP2PAddr(3030)
	require.NoError(t, err)
	require.Equal(t, "/ip4/5.6.7.8/tcp/3030", addr)
}

func TestGetIPWithExternal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ip":"1.2.3.4"}`))
	}))
	defer srv.Close()

	t.Setenv("REMOTE_ADDR", strings.TrimPrefix(srv.URL, "http://"))
	ip, err := netutil.GetIPWithExternal()
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", ip)
}
