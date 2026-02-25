package net_test

import (
	"fmt"
	stdnet "net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	netpkg "github.com/getoptimum/optimum-common/pkg/net"
	"github.com/stretchr/testify/require"
)

func TestGetOutboundQUICP2PAddr(t *testing.T) {
	addr, err := netpkg.GetOutboundQUICP2PAddr(3030)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(addr, "/ip4/"))
	require.True(t, strings.HasSuffix(addr, "/quic-v1"))
}

func TestGetOutboundTCPP2PAddr(t *testing.T) {
	addr, err := netpkg.GetOutboundTCPP2PAddr(3030)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(addr, "/ip4/"))
}

func TestGetIPWithExternal(t *testing.T) {
	ip, err := netpkg.GetOutboundIP()
	require.NoError(t, err)
	require.NotEmpty(t, ip)
}

func TestGetPrivateIPs_Unit(t *testing.T) {
	defer func() {
		netpkg.GetInterfaces = stdnet.Interfaces
		netpkg.ListAddrs = func(iface stdnet.Interface) ([]stdnet.Addr, error) {
			return iface.Addrs()
		}
	}()

	// fake interface list (only 1 for testing)
	netpkg.GetInterfaces = func() ([]stdnet.Interface, error) {
		return []stdnet.Interface{{Index: 1, Name: "mock0"}}, nil
	}

	// fake addresses for that interface
	// in case real net.Interfaces is called mock address will be appended for every interface
	netpkg.ListAddrs = func(iface stdnet.Interface) ([]stdnet.Addr, error) {
		return []stdnet.Addr{
			&stdnet.IPNet{IP: stdnet.ParseIP("10.1.2.3"), Mask: stdnet.CIDRMask(32, 32)},
		}, nil
	}

	ips, err := netpkg.GetPrivateIPs()
	require.NoError(t, err)
	require.Equal(t, []string{"10.1.2.3"}, ips)
}

// helper: collect the same candidate set that ipAddresses() would discover,
// but without relying on internal sorting or mocking
func systemCandidateIPs(t *testing.T) []stdnet.IP {
	t.Helper()

	ifaces, err := stdnet.Interfaces()
	// If os can't list interfaces, tests that call ExternalIP will fail too
	require.NoError(t, err, "net.Interfaces() failed")

	var out []stdnet.IP
	for _, iface := range ifaces {
		// must be up, skip loopback iface
		if iface.Flags&stdnet.FlagUp == 0 || iface.Flags&stdnet.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			// skipping problematic interfaces instead of failing
			continue
		}
		for _, a := range addrs {
			var ip stdnet.IP
			switch v := a.(type) {
			case *stdnet.IPNet:
				ip = v.IP
			case *stdnet.IPAddr:
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
	ipStr, err := netpkg.ExternalIP()
	require.NoError(t, err, "ExternalIP() failed")
	require.NotEmpty(t, ipStr, "ExternalIP() returned empty string")
	ip := stdnet.ParseIP(ipStr)
	require.NotNil(t, ip, "ExternalIP() returned invalid IP: %q", ipStr)
	if ip.IsLoopback() {
		require.Equal(t, "127.0.0.1", ip.String(), "unexpected loopback address")
	}
}

func TestExternalIP_RespectsSystemInterfaces(t *testing.T) {
	gotStr, err := netpkg.ExternalIP()
	require.NoError(t, err, "ExternalIP() should not return error")

	require.NotEmpty(t, gotStr, "ExternalIP() must not return empty string")

	got := stdnet.ParseIP(gotStr)
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
		_, err := netpkg.ExternalIP()
		require.NoError(b, err, "ExternalIP() should not return an error")
	}
}

func TestSortAddresses(t *testing.T) {
	testCases := []struct {
		name     string
		input    []stdnet.IP
		expected []stdnet.IP
	}{
		{
			name:     "IPv4 before IPv6",
			input:    []stdnet.IP{stdnet.ParseIP("2001:db8::1"), stdnet.ParseIP("192.168.1.1")},
			expected: []stdnet.IP{stdnet.ParseIP("192.168.1.1"), stdnet.ParseIP("2001:db8::1")},
		},
		{
			name:     "all IPv4",
			input:    []stdnet.IP{stdnet.ParseIP("10.0.0.1"), stdnet.ParseIP("192.168.1.1")},
			expected: []stdnet.IP{stdnet.ParseIP("10.0.0.1"), stdnet.ParseIP("192.168.1.1")},
		},
		{
			name:     "all IPv6",
			input:    []stdnet.IP{stdnet.ParseIP("2001:db8::2"), stdnet.ParseIP("2001:db8::1")},
			expected: []stdnet.IP{stdnet.ParseIP("2001:db8::2"), stdnet.ParseIP("2001:db8::1")},
		},
		{
			name:     "empty slice",
			input:    []stdnet.IP{},
			expected: []stdnet.IP{},
		},
		{
			name:     "mixed order",
			input:    []stdnet.IP{stdnet.ParseIP("2001:db8::1"), stdnet.ParseIP("10.0.0.1"), stdnet.ParseIP("2001:db8::2"), stdnet.ParseIP("192.168.1.1")},
			expected: []stdnet.IP{stdnet.ParseIP("10.0.0.1"), stdnet.ParseIP("192.168.1.1"), stdnet.ParseIP("2001:db8::1"), stdnet.ParseIP("2001:db8::2")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := netpkg.SortAddresses(tc.input)
			require.Equal(t, len(tc.expected), len(result))
			for i := range tc.expected {
				require.True(t, tc.expected[i].Equal(result[i]), "mismatch at index %d: expected %s, got %s", i, tc.expected[i], result[i])
			}
		})
	}
}

func testIPFunction(t *testing.T, fn func(stdnet.IP) bool, testCases []struct {
	name     string
	ip       string
	expected bool
}) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ip := stdnet.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP: %s", tc.ip)
			result := fn(ip)
			require.Equal(t, tc.expected, result, "IP %s", tc.ip)
		})
	}
}

func TestIsPrivateOrULA(t *testing.T) {
	testIPFunction(t, netpkg.IsPrivateOrULA, []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "IPv4 private 10.0.0.0/8",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "IPv4 private 172.16.0.0/12",
			ip:       "172.16.0.1",
			expected: true,
		},
		{
			name:     "IPv4 private 172.31.0.1",
			ip:       "172.31.0.1",
			expected: true,
		},
		{
			name:     "IPv4 private 192.168.0.0/16",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "IPv4 public",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "IPv6 ULA fc00::1",
			ip:       "fc00::1",
			expected: true,
		},
		{
			name:     "IPv6 ULA fd00::1",
			ip:       "fd00::1",
			expected: true,
		},
		{
			name:     "IPv6 public",
			ip:       "2001:db8::1",
			expected: false,
		},
		{
			name:     "loopback",
			ip:       "127.0.0.1",
			expected: false,
		},
	})
}

func TestIsGlobalUnicast(t *testing.T) {
	testIPFunction(t, netpkg.IsGlobalUnicast, []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "IPv4 public",
			ip:       "8.8.8.8",
			expected: true,
		},
		{
			name:     "IPv4 private",
			ip:       "192.168.1.1",
			expected: false,
		},
		{
			name:     "IPv4 loopback",
			ip:       "127.0.0.1",
			expected: false,
		},
		{
			name:     "IPv4 link-local",
			ip:       "169.254.0.1",
			expected: false,
		},
		{
			name:     "IPv4 multicast",
			ip:       "224.0.0.1",
			expected: false,
		},
		{
			name:     "IPv6 public",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "IPv6 ULA",
			ip:       "fc00::1",
			expected: false,
		},
		{
			name:     "IPv6 link-local",
			ip:       "fe80::1",
			expected: false,
		},
		{
			name:     "IPv6 multicast",
			ip:       "ff02::1",
			expected: false,
		},
	})
}

func TestGetInterfaceIPs(t *testing.T) {
	ips := netpkg.GetInterfaceIPs()
	// Should return at least loopback or some interface IPs
	// We can't predict exact values, but should not panic and should return valid IPs
	for _, ipStr := range ips {
		ip := stdnet.ParseIP(ipStr)
		require.NotNil(t, ip, "invalid IP returned: %s", ipStr)
		require.NotNil(t, ip.To4(), "should return IPv4 addresses only: %s", ipStr)
		require.False(t, ip.IsLoopback(), "should not return loopback: %s", ipStr)
	}
}

func withOverrides(bootstrapURL string, extServices []string, fn func()) {
	origBootstrap := netpkg.BootstrapURL
	origServices := netpkg.ExternalIPServices
	defer func() {
		netpkg.BootstrapURL = origBootstrap
		netpkg.ExternalIPServices = origServices
	}()
	netpkg.BootstrapURL = bootstrapURL
	netpkg.ExternalIPServices = extServices
	fn()
}

func TestGetOutboundIP_BootstrapPublicIP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"ip":"35.200.1.1"}`)
	}))
	defer srv.Close()

	withOverrides(srv.URL, nil, func() {
		ip, err := netpkg.GetOutboundIP()
		require.NoError(t, err)
		require.Equal(t, "35.200.1.1", ip)
	})
}

func TestGetOutboundIP_BootstrapPrivateIP_FallsBackToExternal(t *testing.T) {
	bootstrap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"ip":"10.0.0.13"}`)
	}))
	defer bootstrap.Close()

	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "  203.0.113.50\n")
	}))
	defer external.Close()

	withOverrides(bootstrap.URL, []string{external.URL}, func() {
		ip, err := netpkg.GetOutboundIP()
		require.NoError(t, err)
		require.Equal(t, "203.0.113.50", ip)
	})
}

func TestGetOutboundIP_BootstrapDown_FallsBackToExternal(t *testing.T) {
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "198.51.100.1")
	}))
	defer external.Close()

	withOverrides("http://127.0.0.1:1", []string{external.URL}, func() {
		ip, err := netpkg.GetOutboundIP()
		require.NoError(t, err)
		require.Equal(t, "198.51.100.1", ip)
	})
}

func TestGetOutboundIP_AllExternalReturnPrivate_FallsBackToUDP(t *testing.T) {
	bootstrap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"ip":"172.16.0.5"}`)
	}))
	defer bootstrap.Close()

	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "10.0.0.1")
	}))
	defer external.Close()

	withOverrides(bootstrap.URL, []string{external.URL}, func() {
		ip, err := netpkg.GetOutboundIP()
		require.NoError(t, err)
		require.NotEmpty(t, ip)
		parsed := stdnet.ParseIP(ip)
		require.NotNil(t, parsed)
	})
}

func TestGetOutboundIP_ExternalSkipsBadResponses(t *testing.T) {
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badSrv.Close()

	goodSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "93.184.216.34")
	}))
	defer goodSrv.Close()

	withOverrides("http://127.0.0.1:1", []string{badSrv.URL, goodSrv.URL}, func() {
		ip, err := netpkg.GetOutboundIP()
		require.NoError(t, err)
		require.Equal(t, "93.184.216.34", ip)
	})
}
