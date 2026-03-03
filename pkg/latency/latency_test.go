package latency_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/latency"
	"github.com/stretchr/testify/require"
)

func TestLatencyMsInRange(t *testing.T) {
	tests := []struct {
		name      string
		latencyMs int64
		maxMs     int64
		wantOk    bool
	}{
		{"zero_ok", 0, 60_000, true},
		{"in_range", 100, 60_000, true},
		{"max_inclusive", 60_000, 60_000, true},
		{"negative_latency", -1, 60_000, false},
		{"negative_max", 100, -1, false},
		{"over_max", 61_000, 60_000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := latency.LatencyMsInRange(tt.latencyMs, tt.maxMs)
			require.Equal(t, tt.latencyMs, got)
			require.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestRecvTimestampReasonable(t *testing.T) {
	now := int64(1_700_000_000_000) // ms

	tests := []struct {
		name         string
		recvAtMs     int64
		nowMs        int64
		maxFutureMs  int64
		wantReasonable bool
	}{
		{"now_ok", now, now, 5000, true},
		{"past_ok", now - 1000, now, 5000, true},
		{"zero_bad", 0, now, 5000, false},
		{"negative_bad", -1, now, 5000, false},
		{"slightly_future_ok", now + 100, now, 5000, true},
		{"too_far_future_bad", now + 10_000, now, 5000, false},
		{"negative_max_future_treated_as_zero", now + 1, now, -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := latency.RecvTimestampReasonable(tt.recvAtMs, tt.nowMs, tt.maxFutureMs)
			require.Equal(t, tt.wantReasonable, got)
		})
	}
}
