package shm

import "time"

// Phases of one shared-memory operation, after the caller is admitted. wait is
// the only phase measuring the rlnc-server rather than the calling process.
const (
	PhaseAdmission = "admission"
	PhaseClaim     = "claim"
	PhaseSubmit    = "submit"
	PhaseWait      = "wait"
	PhaseRead      = "read"
)

// PhaseServer attributes a failure to the rlnc-server rejecting the operation.
// It is a failure cause only, never a timed phase. The client read the output
// frame fine, the server just reported an error inside it. Folding it into read
// or wait would make those phases' failure counts unusable.
const PhaseServer = "server"

// Outcomes of one shared-memory operation, as reported end-to-end.
const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
)

// Observer allows callers to observe lane occupancy and per-phase timings of
// shared-memory operations, for metrics.
type Observer interface {
	// AdmissionWait reports time spent waiting for a free lane, which is
	// back-pressure rather than work.
	AdmissionWait(op string, d time.Duration)

	// Phase reports how long one phase of an admitted operation took.
	Phase(op, phase string, d time.Duration)

	// Operation reports the end-to-end latency of a single operation, including
	// its admission wait. Phase quantiles cannot be summed into this, so it is
	// reported separately rather than derived.
	Operation(op, outcome string, d time.Duration)

	// Failure reports an operation that ended in the named phase.
	Failure(op, phase string)

	// LaneClaimed and LaneReleased are paired around a held lane; one without
	// the other leaves an occupancy gauge drifting forever.
	LaneClaimed()
	LaneReleased()
}
