package utils_test

import (
	"net"
	"strings"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestGetOutboundQUICP2PAddr(t *testing.T) {
	addr, err := utils.GetOutboundQUICP2PAddr(3030)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(addr, "/ip4/"))
	require.True(t, strings.HasSuffix(addr, "/quic-v1"))
}

func TestGetOutboundTCPP2PAddr(t *testing.T) {
	addr, err := utils.GetOutboundTCPP2PAddr(3030)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(addr, "/ip4/"))
}

func TestGetIPWithExternal(t *testing.T) {
	ip, err := utils.GetOutboundIP()
	require.NoError(t, err)
	require.NotEmpty(t, ip)
}

func TestGetPrivateIPs_Unit(t *testing.T) {
	defer func() {
		utils.GetInterfaces = net.Interfaces
		utils.ListAddrs = func(iface net.Interface) ([]net.Addr, error) {
			return iface.Addrs()
		}
	}()

	// fake interface list (only 1 for testing)
	utils.GetInterfaces = func() ([]net.Interface, error) {
		return []net.Interface{{Index: 1, Name: "mock0"}}, nil
	}

	// fake addresses for that interface
	// in case real net.Interfaces is called mock address will be appended for every interface
	utils.ListAddrs = func(iface net.Interface) ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("10.1.2.3"), Mask: net.CIDRMask(32, 32)},
		}, nil
	}

	ips, err := utils.GetPrivateIPs()
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
	ipStr, err := utils.ExternalIP()
	require.NoError(t, err, "ExternalIP() failed")
	require.NotEmpty(t, ipStr, "ExternalIP() returned empty string")
	ip := net.ParseIP(ipStr)
	require.NotNil(t, ip, "ExternalIP() returned invalid IP: %q", ipStr)
	if ip.IsLoopback() {
		require.Equal(t, "127.0.0.1", ip.String(), "unexpected loopback address")
	}
}

func TestExternalIP_RespectsSystemInterfaces(t *testing.T) {
	gotStr, err := utils.ExternalIP()
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
		_, err := utils.ExternalIP()
		require.NoError(b, err, "ExternalIP() should not return an error")
	}
}
