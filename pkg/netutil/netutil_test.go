package netutil_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/netutil"
	"github.com/stretchr/testify/require"
)

func TestGetOutboundP2PAddr(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify request method
		require.Equal(t, http.MethodGet, r.Method, "expected GET request")

		// verify endpoint
		require.Equal(t, "/return_ip", r.URL.Path, "wrong path used")

		called = true

		_, _ = w.Write([]byte(`{"ip":"5.6.7.8"}`))
	}))
	defer srv.Close()

	t.Setenv("REMOTE_HOST", strings.TrimPrefix(srv.URL, "http://"))

	addr, err := netutil.GetOutboundP2PAddr(3030)
	require.NoError(t, err)
	require.Equal(t, "/ip4/5.6.7.8/tcp/3030", addr)
	require.True(t, called, "expected handler to be called")
}

func TestGetIPWithExternal(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify request method and endpoint
		require.Equal(t, http.MethodGet, r.Method, "expected GET request")
		require.Equal(t, "/return_ip", r.URL.Path, "wrong path used")

		called = true

		_, _ = w.Write([]byte(`{"ip":"1.2.3.4"}`))
	}))
	defer srv.Close()

	// Point REMOTE_ADDR to the test server (without http:// prefix)
	t.Setenv("REMOTE_ADDR", strings.TrimPrefix(srv.URL, "http://"))

	ip, err := netutil.GetIPWithExternal()
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", ip)
	require.True(t, called, "expected handler to be called")
}

func TestGetPrivateIPs_Unit(t *testing.T) {
	defer func() {
		netutil.GetInterfaces = net.Interfaces
		netutil.ListAddrs = func(iface net.Interface) ([]net.Addr, error) {
			return iface.Addrs()
		}
	}()

	// fake interface list (only 1 for testing)
	netutil.GetInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{{Index: 1, Name: "mock0"}}, nil
	}

	// fake addresses for that interface
	// in case real net.Interfaces is called mock adress will be appended for every interface
	netutil.ListAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("10.1.2.3"), Mask: net.CIDRMask(32, 32)},
		}, nil
	}

	ips, err := netutil.GetPrivateIPs()
	require.NoError(t, err)
	require.Equal(t, []string{"10.1.2.3"}, ips)
}
