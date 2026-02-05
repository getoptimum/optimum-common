package net_test

import (
	stdnet "net"
	"testing"

	netpkg "github.com/getoptimum/optimum-common/pkg/net"
	"github.com/stretchr/testify/require"
)

// FuzzMultiAddressBuilder tests multiaddress construction
func FuzzMultiAddressBuilder(f *testing.F) {
	f.Add(byte(192), byte(168), byte(1), byte(1), 4001)
	f.Add(byte(10), byte(0), byte(0), byte(1), 8080)
	f.Add(byte(127), byte(0), byte(0), byte(1), 0)
	f.Add(byte(8), byte(8), byte(8), byte(8), 65535)

	f.Fuzz(func(t *testing.T, a, b, c, d byte, port int) {
		ip := stdnet.IPv4(a, b, c, d)
		addrs, err := netpkg.MultiAddressBuilder(ip, port)
		if err == nil {
			require.NotEmpty(t, addrs, "expected at least one multiaddr on success")
		}
	})
}
