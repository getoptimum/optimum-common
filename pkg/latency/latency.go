package latency

// LatencyMsInRange reports whether latencyMs is in [0, maxMs] (inclusive).
// Used to avoid recording negative (clock skew) or absurdly large (clock jump) values.
// Returns (latencyMs, true) when valid; (latencyMs, false) when out of range.
func LatencyMsInRange(latencyMs, maxMs int64) (int64, bool) {
	if latencyMs < 0 || maxMs < 0 || latencyMs > maxMs {
		return latencyMs, false
	}
	return latencyMs, true
}

// RecvTimestampReasonable reports whether recvAtMs is plausible relative to nowMs.
// Returns false if recvAt is in the future by more than maxFutureMs (clock skew),
// or if recvAt is zero/negative. Caller passes time.Now().UnixMilli() as nowMs.
func RecvTimestampReasonable(recvAtMs, nowMs, maxFutureMs int64) bool {
	if recvAtMs <= 0 {
		return false
	}
	if maxFutureMs < 0 {
		maxFutureMs = 0
	}
	return recvAtMs <= nowMs+maxFutureMs
}
