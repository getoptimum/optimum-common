package net_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	netpkg "github.com/getoptimum/optimum-common/pkg/net"
)

func BenchmarkExternalIP(b *testing.B) {
	for b.Loop() {
		_, err := netpkg.ExternalIP()
		require.NoError(b, err, "ExternalIP() should not return an error")
	}
}
