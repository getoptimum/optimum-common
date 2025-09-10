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
	// in case real net.Interfaces is called mock address will be appended for every interface
	netutil.ListAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("10.1.2.3"), Mask: net.CIDRMask(32, 32)},
		}, nil
	}

	ips, err := netutil.GetPrivateIPs()
	require.NoError(t, err)
	require.Equal(t, []string{"10.1.2.3"}, ips)
}

// helper: collect the same candidate set that ipAddresses() would discover,
// but without relying on internal sorting or mocking
func systemCandidateIPs(t *testing.T) []net.IP {
	t.Helper()

	ifaces, err := net.Interfaces()
	// If os can't list interfaces, tests that call ExternalIP will fail too
	require.NoError(t, err, "net.Interfaces() failed")

	var out []net.IP
	for _, iface := range ifaces {
		// must be up, skip loopback iface
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			// skipping problematic interfaces instead of failing
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			// skip loopback and link-local
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			out = append(out, ip)
		}
	}
	return out
}

func TestExternalIP_ReturnsValidIP(t *testing.T) {
	ipStr, err := netutil.ExternalIP()
	require.NoError(t, err, "ExternalIP() failed")
	require.NotEmpty(t, ipStr, "ExternalIP() returned empty string")
	ip := net.ParseIP(ipStr)
	require.NotNil(t, ip, "ExternalIP() returned invalid IP: %q", ipStr)
	if ip.IsLoopback() {
		require.Equal(t, "127.0.0.1", ip.String(), "unexpected loopback address")
	}
}

func TestExternalIP_RespectsSystemInterfaces(t *testing.T) {
	gotStr, err := netutil.ExternalIP()
	require.NoError(t, err, "ExternalIP() should not return error")

	require.NotEmpty(t, gotStr, "ExternalIP() must not return empty string")

	got := net.ParseIP(gotStr)
	require.NotNil(t, got, "ExternalIP() returned invalid IP: %q", gotStr)

	candidates := systemCandidateIPs(t)

	if len(candidates) == 0 {
		// No non-loopback, non–link-local addresses on this host
		// must fallback to 127.0.0.1.
		require.Equal(t, "127.0.0.1", gotStr,
			"Expected fallback 127.0.0.1 on host with no external IPs")
		return
	}

	// otherwise, the returned IP must be one of the candidates.
	found := false
	for _, c := range candidates {
		if c.Equal(got) {
			found = true
			break
		}
	}
	require.Truef(t, found,
		"ExternalIP()=%s not among system candidates %v", gotStr, candidates)

	require.Falsef(t, got.IsLoopback(),
		"ExternalIP() returned loopback despite having external candidates: %s", gotStr)
}

func BenchmarkExternalIP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := netutil.ExternalIP()
		require.NoError(b, err, "ExternalIP() should not return an error")
	}
}
