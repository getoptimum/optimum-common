package net_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/getoptimum/optimum-common/pkg/net"
)

func TestBuildAdvertisedQUICAddresses(t *testing.T) {
	l := logger.NewAppSLogger(logger.Debug)

	t.Run("valid payload", func(t *testing.T) {
		res, err := net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "2001:db8::1", 4001)
		require.NoError(t, err)

		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/udp/4001/quic-v1"))
		require.Contains(t, res, mustMA(t, "/ip6/2001:db8::1/udp/4001/quic-v1"))
	})
	t.Run("should return err on invalid ipv4", func(t *testing.T) {
		res, err := net.BuildAdvertisedQUICAddresses(l, "not-an-ip", "2001:db8::1", 4001)
		require.Error(t, err)
		require.Empty(t, res)
	})
	t.Run("should work with empty ipv6", func(t *testing.T) {
		res, err := net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "", 4001)
		require.NoError(t, err)
		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/udp/4001/quic-v1"))
	})
	t.Run("should not fail on invalid ipv6 and still return ipv4", func(t *testing.T) {
		res, err := net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "not-an-ip", 4001)
		require.NoError(t, err)
		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/udp/4001/quic-v1"))
	})
	t.Run("should return err on invalid port", func(t *testing.T) {
		res, err := net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "2001:db8::1", 0)
		require.Error(t, err)
		require.Empty(t, res)

		res, err = net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "2001:db8::1", 70000)
		require.Error(t, err)
		require.Empty(t, res)
	})
	t.Run("should accept boundary ports", func(t *testing.T) {
		res, err := net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "", 1)
		require.NoError(t, err)
		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/udp/1/quic-v1"))

		res, err = net.BuildAdvertisedQUICAddresses(l, "1.2.3.4", "", 65535)
		require.NoError(t, err)
		require.Contains(t, res, mustMA(t, "/ip4/1.2.3.4/udp/65535/quic-v1"))
	})
}
