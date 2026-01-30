package utils_test

import (
	"net"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
)

// FuzzMultiaddr tests multiaddress parsing and IP protocol detection
func FuzzMultiaddr(f *testing.F) {
	// Valid peer JSON
	f.Add(`{"peerID":"QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N","addrs":[]}`, "192.168.1.1")
	f.Add(`{"peerID":"12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN","addrs":["/ip4/127.0.0.1/tcp/4001"]}`, "10.0.0.1")
	f.Add(`{}`, "::1")
	f.Add(`{"peerID":"","addrs":[]}`, "2001:db8::1")
	f.Add("", "8.8.8.8")
	f.Add("{", "not-an-ip")
	f.Add(`null`, "")
	f.Add(`{"peerID":"invalid!@#","addrs":[]}`, "256.256.256.256")

	f.Fuzz(func(t *testing.T, peerJSON string, ipStr string) {
		// Test AddressInfoFromString - must not panic
		_, _ = utils.AddressInfoFromString(peerJSON)

		// Test GetIPProtocol - must return "ip4" or "ip6"
		proto := utils.GetIPProtocol(ipStr)
		if proto != "ip4" && proto != "ip6" {
			t.Fatalf("GetIPProtocol: expected ip4 or ip6, got %q", proto)
		}
	})
}

// FuzzMultiAddressBuilder tests multiaddress construction
func FuzzMultiAddressBuilder(f *testing.F) {
	f.Add(byte(192), byte(168), byte(1), byte(1), 4001)
	f.Add(byte(10), byte(0), byte(0), byte(1), 8080)
	f.Add(byte(127), byte(0), byte(0), byte(1), 0)
	f.Add(byte(8), byte(8), byte(8), byte(8), 65535)

	f.Fuzz(func(t *testing.T, a, b, c, d byte, port int) {
		ip := net.IPv4(a, b, c, d)
		addrs, err := utils.MultiAddressBuilder(ip, port)
		if err == nil && len(addrs) == 0 {
			t.Fatal("expected at least one multiaddr on success")
		}
	})
}
