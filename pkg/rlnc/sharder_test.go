package rlnc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/rlnc"
)

func TestShardCount(t *testing.T) {
	testCases := []struct {
		name         string
		payloadLen   uint64
		maxShardSize uint32
		output       uint64
	}{
		{name: "zero max shard size", payloadLen: 100, maxShardSize: 0, output: 0},
		{name: "fits one shard", payloadLen: 84, maxShardSize: 100, output: 1},
		{name: "metadata spills into second shard", payloadLen: 85, maxShardSize: 100, output: 2},
		{name: "exact multiple", payloadLen: 184, maxShardSize: 100, output: 2},
		{name: "empty payload", payloadLen: 0, maxShardSize: 8, output: 2},
		{name: "large payload", payloadLen: 1 << 20, maxShardSize: 8000, output: 132},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.output, rlnc.ShardCount(tc.payloadLen, tc.maxShardSize))
		})
	}
}
