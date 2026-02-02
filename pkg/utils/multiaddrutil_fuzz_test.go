package utils_test

import (
	"net"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

// FuzzMultiAddressBuilder tests multiaddress construction
func FuzzMultiAddressBuilder(f *testing.F) {
	f.Add(byte(192), byte(168), byte(1), byte(1), 4001)
	f.Add(byte(10), byte(0), byte(0), byte(1), 8080)
	f.Add(byte(127), byte(0), byte(0), byte(1), 0)
	f.Add(byte(8), byte(8), byte(8), byte(8), 65535)

	f.Fuzz(func(t *testing.T, a, b, c, d byte, port int) {
		ip := net.IPv4(a, b, c, d)
		addrs, err := utils.MultiAddressBuilder(ip, port)
		if err == nil {
			require.NotEmpty(t, addrs, "expected at least one multiaddr on success")
		}
	})
}
