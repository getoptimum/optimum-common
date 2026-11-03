package shm

import "github.com/getoptimum/shm/pkg/shm"

// clientLane exposes the client-side view of a shared-memory lane.
type clientLane struct {
	mem *shm.Mem
}

// newClientLane returns a client-side lane wrapper for mem.
func newClientLane(mem *shm.Mem) clientLane {
	return clientLane{mem: mem}
}

// ReceiveResponse reads a response frame from the lane's output region.
func (l clientLane) ReceiveResponse() ([]byte, error) {
	return l.mem.ReadOutputFrame()
}

// ClearFrames clears the visible request and response frames for reuse.
func (l clientLane) ClearFrames() {
	l.mem.ClearInputFrameHeader()
	l.mem.ClearOutputFrameHeader()
}
