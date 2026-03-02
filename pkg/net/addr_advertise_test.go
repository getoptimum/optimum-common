package net_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/getoptimum/optimum-common/pkg/net"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestBuildAdvertisedAddresses(t *testing.T) {
	l := logger.NewAppSLogger(logger.Debug)

	t.Run("valid payload", func(t *testing.T) {
		// when
		res, err := net.BuildAdvertisedAddresses(l, "1.2.3.4", "2001:db8::1", 4001)
		require.NoError(t, err)

		// then
		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/tcp/4001"))
		require.Contains(t, res, mustMA(t, "/ip6/2001:db8::1/tcp/4001"))
	})
	t.Run("should return err on invalid ipv4", func(t *testing.T) {
		// when, then
		res, err := net.BuildAdvertisedAddresses(l, "not-an-ip", "2001:db8::1", 4001)
		require.Error(t, err)
		require.Empty(t, res)
	})
	t.Run("should not fail on invalid ipv6 and still return ipv4", func(t *testing.T) {
		// when
		res, err := net.BuildAdvertisedAddresses(l, "1.2.3.4", "not-an-ip", 4001)
		require.NoError(t, err)

		// then
		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/tcp/4001"))
	})
	t.Run("should return err on invalid port", func(t *testing.T) {
		// when, then
		res, err := net.BuildAdvertisedAddresses(l, "1.2.3.4", "2001:db8::1", 0)
		require.Error(t, err)
		require.Empty(t, res)

		res, err = net.BuildAdvertisedAddresses(l, "1.2.3.4", "2001:db8::1", 70000)
		require.Error(t, err)
		require.Empty(t, res)
	})
}

func mustMA(t *testing.T, s string) multiaddr.Multiaddr {
	t.Helper()
	ma, err := multiaddr.NewMultiaddr(s)
	require.NoError(t, err)
	return ma
}
