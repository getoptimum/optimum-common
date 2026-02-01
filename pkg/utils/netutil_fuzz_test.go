package utils_test

import (
	"net"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

// FuzzIPClassification tests IP address classification functions
func FuzzIPClassification(f *testing.F) {
	f.Add(byte(10), byte(0), byte(0), byte(1))       // Private
	f.Add(byte(192), byte(168), byte(1), byte(1))    // Private
	f.Add(byte(8), byte(8), byte(8), byte(8))        // Public
	f.Add(byte(127), byte(0), byte(0), byte(1))      // Loopback
	f.Add(byte(169), byte(254), byte(1), byte(1))    // Link-local
	f.Add(byte(224), byte(0), byte(0), byte(1))      // Multicast
	f.Add(byte(172), byte(16), byte(0), byte(1))     // Private 172.16.0.0/12
	f.Add(byte(0), byte(0), byte(0), byte(0))        // Unspecified

	f.Fuzz(func(t *testing.T, a, b, c, d byte) {
		ip := net.IPv4(a, b, c, d)

		isPrivate := utils.IsPrivateOrULA(ip)
		isGlobal := utils.IsGlobalUnicast(ip)

		// Cannot be both private/ULA AND global unicast
		require.False(t, isPrivate && isGlobal, "IP %v is both private and global", ip)
	})
}

// FuzzSortAddresses tests address sorting
func FuzzSortAddresses(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(3), uint8(2))
	f.Add(uint8(10), uint8(5))

	f.Fuzz(func(t *testing.T, ipv4Count, ipv6Count uint8) {
		// Cap to avoid excessive memory
		v4 := int(ipv4Count % 20)
		v6 := int(ipv6Count % 20)

		ips := make([]net.IP, 0, v4+v6)
		for i := 0; i < v4; i++ {
			ips = append(ips, net.IPv4(byte(i), byte(i), byte(i), byte(i)))
		}
		for i := 0; i < v6; i++ {
			ips = append(ips, net.ParseIP("2001:db8::1"))
		}

		result := utils.SortAddresses(ips)

		require.Equal(t, len(ips), len(result), "length changed: %d -> %d", len(ips), len(result))

		// IPv4 must come before IPv6
		seenIPv6 := false
		for _, ip := range result {
			if ip.To4() != nil {
				require.False(t, seenIPv6, "IPv4 found after IPv6")
			} else {
				seenIPv6 = true
			}
		}
	})
}
