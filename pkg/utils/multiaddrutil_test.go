package utils_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestMultiAddressBuilder(t *testing.T) {
	port := 30303
	cases := map[string]net.IP{
		"/ip4/192.152.1.3/tcp/30303":                          net.ParseIP("192.152.1.3"),
		"/ip4/127.0.0.1/tcp/30303":                            net.ParseIP("127.0.0.1"),
		"/ip6/::1/tcp/30303":                                  net.ParseIP("::1"),
		"/ip6/2001:db8:11a3:9d7:1f34:8a2e:7a0:765d/tcp/30303": net.ParseIP("2001:db8:11a3:9d7:1f34:8a2e:7a0:765d"),
	}

	for expected, addr := range cases {
		t.Run(expected, func(t *testing.T) {
			res, err := utils.MultiAddressBuilder(addr, port)
			require.NoError(t, err)
			require.Len(t, res, 1)
			require.Equal(t, expected, res[0].String())
		})
	}
}

func TestMultiAddressBuilder_InvalidIP(t *testing.T) {
	_, err := utils.MultiAddressBuilder(net.IP{}, 30303)
	require.Error(t, err)
	require.Contains(t, err.Error(), "neither IPv4 nor IPv6")
}

func TestAddressInfoToStringAndBack(t *testing.T) {
	maList := make([]ma.Multiaddr, 0, 10)
	for i := 30300; i < 30310; i++ {
		m, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", i))
		require.NoError(t, err)
		maList = append(maList, m)
	}

	original := peer.AddrInfo{
		ID:    "16Uiu2HAkxohm96jeTn18K7iKVdNDd5A2jgY1pFoj6SkegPFq6Tzb",
		Addrs: maList,
	}
	encoded := utils.AddressInfoToString(original)

	decoded, err := utils.AddressInfoFromString(encoded)
	require.NoError(t, err)

	require.Equal(t, original.ID, decoded.ID)
	require.Len(t, decoded.Addrs, len(original.Addrs))
	for i := range original.Addrs {
		require.Equal(t, original.Addrs[i].String(), decoded.Addrs[i].String())
	}
}

func TestAddressInfoFromString_InvalidID(t *testing.T) {
	// Create invalid base58 ID
	jsonStr := `{"peerID": "not-a-base58", "addrs": []}`
	_, err := utils.AddressInfoFromString(jsonStr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse peer ID")
}

func TestAddressInfoFromString_InvalidJSON(t *testing.T) {
	_, err := utils.AddressInfoFromString(`{invalid json}`)
	require.Error(t, err)
}

func TestGetIPProtocol(t *testing.T) {
	testCases := []struct {
		name     string
		ipStr    string
		expected string
	}{
		{
			name:     "IPv4 address",
			ipStr:    "192.168.1.1",
			expected: "ip4",
		},
		{
			name:     "IPv6 address",
			ipStr:    "2001:db8::1",
			expected: "ip6",
		},
		{
			name:     "IPv4 loopback",
			ipStr:    "127.0.0.1",
			expected: "ip4",
		},
		{
			name:     "IPv6 loopback",
			ipStr:    "::1",
			expected: "ip6",
		},
		{
			name:     "invalid IP defaults to ip4",
			ipStr:    "invalid",
			expected: "ip4",
		},
		{
			name:     "empty string defaults to ip4",
			ipStr:    "",
			expected: "ip4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.GetIPProtocol(tc.ipStr)
			require.Equal(t, tc.expected, result)
		})
	}
}
