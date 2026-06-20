package lane

import "github.com/getoptimum/shm/pkg/shm"

// ClientLane exposes the client-side view of a shared-memory lane.
type ClientLane struct {
	mem *shm.Mem
}

// NewClientLane returns a client-side lane wrapper for mem.
func NewClientLane(mem *shm.Mem) ClientLane {
	return ClientLane{mem: mem}
}

// SendRequest writes a request frame into the lane's input region.
func (l ClientLane) SendRequest(payload []byte) error {
	return l.mem.WriteInputFrame(payload)
}

// ReceiveResponse reads a response frame from the lane's output region.
func (l ClientLane) ReceiveResponse() ([]byte, error) {
	return l.mem.ReadOutputFrame()
}

// ClearFrames clears the visible request and response frames for reuse.
func (l ClientLane) ClearFrames() {
	l.mem.ClearInputFrameHeader()
	l.mem.ClearOutputFrameHeader()
}

// ServerLane exposes the server-side view of a shared-memory lane.
type ServerLane struct {
	mem *shm.Mem
}

// NewServerLane returns a server-side lane wrapper for mem.
func NewServerLane(mem *shm.Mem) ServerLane {
	return ServerLane{mem: mem}
}

// ReceiveRequest reads a request frame from the lane's input region.
func (l ServerLane) ReceiveRequest() ([]byte, error) {
	return l.mem.ReadInputFrame()
}

// SendResponse writes a response frame into the lane's output region.
func (l ServerLane) SendResponse(payload []byte) error {
	return l.mem.WriteOutputFrame(payload)
}

// Responder writes an operation response to the transport.
type Responder interface {
	SendResponse([]byte) error
}
