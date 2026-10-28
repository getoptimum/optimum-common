package shm

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"
	"syscall"
	"time"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	mathutil "github.com/getoptimum/optimum-common/pkg/math"
	shmpb "github.com/getoptimum/optimum-common/pkg/rlncpb/shm"
	"github.com/getoptimum/shm/pkg/shm"
)

// shmWaitTimeout is the maximum duration to wait for a lane to process an operation and produce an output frame.
const shmWaitTimeout = 2 * time.Second

// Field numbers of the OpInput and OpOutput envelopes, which this package
// encodes and parses by hand. See the package documentation before changing
// them. TestOpFrameFieldNumbersMatchSchema pins them to the generated schema.
const (
	opInputOperationField    protowire.Number = 1
	opInputMsgField          protowire.Number = 2
	opInputFrameVersionField protowire.Number = 3

	opOutputResponseField protowire.Number = 1
	opOutputErrorField    protowire.Number = 2
)

// FrameVersion is the RLNC frame layout this node expects rlnc-server to
// produce and accept. It must match the server's types.FrameVersion. Every
// OpInput carries it, so a server on another layout rejects each operation.
const FrameVersion uint32 = 1

var (
	// errFrameTooLarge reports a request whose OpInput frame does not fit the
	// lane's input region.
	errFrameTooLarge = errors.New("op input frame exceeds lane input region")

	// ErrIncompatibleServer reports an rlnc-server that predates frame
	// versioning or uses a different frame layout.
	ErrIncompatibleServer = errors.New("incompatible rlnc-server")
)

type (
	// Service defines a wrapper around a shared memory lane set with a set of utility
	// methods to execute RLNC APIs through shared memory.
	Service struct {
		lanes    int
		ls       LaneSet
		nextLane atomic.Uint64
		// slots bounds local callers to the number of physical lanes. Without this
		// admission gate, transient local contention was reported as a hard "no
		// idle lanes" operation failure instead of applying back-pressure.
		slots chan struct{}

		// observer is an optional metric sink.
		observer Observer
	}

	// claimedLane represents a lane that has been claimed for exclusive use by a
	// client. It contains the shared memory region, lane index, and generation
	// number, which is used to ensure that the lane is not released or reused
	// while the client is using it.
	claimedLane struct {
		mem        *shm.Mem
		lane       int
		generation uint64
	}

	// phaseTimer walks an operation's phases, reporting each as it is passed, so
	// execute still reads as the sequence of steps it already was. start is
	// kept so the whole operation can be reported alongside its breakdown.
	phaseTimer struct {
		svc   *Service
		op    string
		start time.Time
		last  time.Time
	}
)

// NewWithLaneSet returns a Service over an existing lane set. Callers are
// admitted up to lanes at a time. Tests use it to inject a stub LaneSet.
func NewWithLaneSet(ls LaneSet, lanes int) *Service {
	return &Service{
		ls:    ls,
		lanes: lanes,
		slots: make(chan struct{}, max(0, lanes)),
	}
}

// New attaches to the lanes an rlnc-server published under name. The server
// must already be running with a matching name and lane count.
func New(lanes int, name string) (*Service, error) {
	ls, err := shm.AttachLaneSet(lanes, name)
	if err != nil {
		return nil, fmt.Errorf("attach lane set: %w", err)
	}

	return &Service{
		lanes: lanes,
		ls:    ls,
		slots: make(chan struct{}, lanes),
	}, nil
}

// VerifyServer asks rlnc-server for its frame version and returns its build
// version. Servers that predate frame versioning reject the operation as
// unsupported. Either case, or a different frame version, wraps
// ErrIncompatibleServer.
func (s *Service) VerifyServer() (serverVersion string, err error) {
	var resp shmpb.ServerInfoResponse
	err = s.ExecuteMessage(shmpb.OperationType_OPERATION_TYPE_SERVER_INFO, &shmpb.ServerInfoRequest{}, &resp)
	if opErr, ok := errors.AsType[*shmpb.OpError](err); ok &&
		opErr.GetCode() == shmpb.OpErrorCode_OP_ERROR_CODE_UNSUPPORTED_OPERATION {
		return "", fmt.Errorf("%w: server predates frame versioning: %w", ErrIncompatibleServer, err)
	}
	if err != nil {
		return "", fmt.Errorf("execute server info request: %w", err)
	}

	if resp.GetFrameVersion() != FrameVersion {
		return resp.GetServerVersion(), fmt.Errorf(
			"%w: frame version %d, expected %d",
			ErrIncompatibleServer,
			resp.GetFrameVersion(),
			FrameVersion,
		)
	}

	return resp.GetServerVersion(), nil
}

// SetObserver installs the metric sink. Call before the service is used.
func (s *Service) SetObserver(o Observer) { s.observer = o }

// Capacity returns the number of shared-memory lanes available to concurrent
// callers.
func (s *Service) Capacity() int {
	if s == nil {
		return 0
	}

	return s.lanes
}

// ExecuteOp waits for local SHM capacity, then executes one operation on an
// exclusively claimed lane. The wait is deliberate back-pressure, i.e. callers
// are bounded by the lane count instead of failing merely because another local
// goroutine is using the service concurrently.
func (s *Service) ExecuteOp(op shmpb.OperationType, payloadBz []byte) ([]byte, error) {
	var response []byte

	err := s.execute(
		op,
		len(payloadBz),
		func(dst []byte) ([]byte, error) { return append(dst, payloadBz...), nil },
		func(src []byte) error {
			response = bytes.Clone(src)
			return nil
		},
	)

	return response, err
}

// ExecuteMessage is ExecuteOp for protobuf messages. It marshals req straight
// into the lane's input frame and unmarshals resp straight out of the output
// frame. A request is never staged in a separate buffer and a response is
// copied once. Decode requests carry a chunk's whole symbol set, so the staging
// copies ExecuteOp's callers make dominate their allocations.
func (s *Service) ExecuteMessage(op shmpb.OperationType, req, resp proto.Message) error {
	return s.execute(
		op,
		proto.Size(req),
		func(dst []byte) ([]byte, error) { return proto.MarshalOptions{}.MarshalAppend(dst, req) },
		func(src []byte) error { return proto.Unmarshal(src, resp) },
	)
}

// execute runs one operation. appendMsg appends exactly msgSize bytes of
// request to dst, and readResponse consumes the response while it still lives
// in the lane, before the lane is released.
func (s *Service) execute(
	op shmpb.OperationType,
	msgSize int,
	appendMsg func(dst []byte) ([]byte, error),
	readResponse func(src []byte) error,
) error {
	timer := s.newPhaseTimer(op.String())

	// Rejected before the admission wait it names could begin, so there is no
	// phase to close. The caller still saw a failed operation, so it is still
	// reported end to end.
	if s.lanes <= 0 || s.slots == nil {
		timer.failBeforeAdmission()

		return fmt.Errorf("lane set is closed")
	}

	// Blocking send: the time spent here is back-pressure, not work.
	s.slots <- struct{}{}
	defer func() { <-s.slots }()

	timer.markAdmission()

	claimed, err := s.claimLane()
	if err != nil {
		timer.fail(PhaseClaim)
		return err
	}

	timer.mark(PhaseClaim)

	if s.observer != nil {
		s.observer.LaneClaimed()

		// Same rule as the phase callbacks: the sink's cost is not submit's.
		timer.resume()
	}

	laneClient := newClientLane(claimed.mem)
	stateAddr := claimed.mem.StateAddr()

	var (
		submitted     bool
		outputClaimed bool
	)

	defer func() {
		laneDirty := !submitted || outputClaimed

		// A lane must not become StateIdle while its in-process mutex is still
		// locked. So we first check if the lane was not submitted or if the output
		// was claimed, and if so, we clear the frames to ensure that the lane is in
		// a consistent state before unlocking it.
		if laneDirty {
			laneClient.ClearFrames()
		}

		// Always unlock the lane, even if the operation failed, to avoid deadlocks.
		_ = s.ls.UnlockLane(claimed.lane) // best effort

		// Finally, if the lane was dirty, we attempt to release the claim and wake
		// up any waiters on the lane's state.
		if laneDirty && claimed.mem.ReleaseClaim(claimed.generation) {
			_ = shm.FutexWake(stateAddr, 1) // best effort
		}

		// Paired with LaneClaimed, and always run, so the gauge cannot drift.
		if s.observer != nil {
			s.observer.LaneReleased()
		}
	}()

	// 1. Write the operation and payload to the lane's input frame.
	if err := writeInputFrame(claimed.mem.Input(), op, msgSize, appendMsg); err != nil {
		timer.fail(PhaseSubmit)
		return fmt.Errorf("write request: %w", err)
	}

	// 2. Signal the lane that the input frame is ready for processing.
	if !claimed.mem.SubmitClaimed(claimed.generation) {
		timer.fail(PhaseSubmit)
		return fmt.Errorf("submit lane %d generation %d", claimed.lane, claimed.generation)
	}

	submitted = true

	if err := shm.FutexWake(stateAddr, 1); err != nil {
		timer.fail(PhaseSubmit)
		return fmt.Errorf("futex wake input ready: %w", err)
	}

	timer.mark(PhaseSubmit)

	// 3. Wait for the lane to process the input frame and write the output frame.
	waitCtx, cancelWait := context.WithTimeout(context.Background(), shmWaitTimeout)
	defer cancelWait()

	if err := waitForState(waitCtx, claimed.mem, shm.StateOutputReady); err != nil {
		timer.fail(PhaseWait)

		return fmt.Errorf("wait output ready: %w", err)
	}

	timer.mark(PhaseWait)

	if !claimed.mem.TryClaimOutput(claimed.generation) {
		timer.fail(PhaseRead)
		return fmt.Errorf("claim output for lane %d generation %d", claimed.lane, claimed.generation)
	}

	outputClaimed = true

	// 4. Read the output frame in place. It aliases the lane, so the response is
	// consumed before the deferred release can hand the lane to someone else.
	opOutputBz, err := laneClient.ReceiveResponse()
	if err != nil {
		timer.fail(PhaseRead)
		return fmt.Errorf("read response: %w", err)
	}

	response, opErr, err := parseOpOutput(opOutputBz)
	if err != nil {
		timer.fail(PhaseRead)
		return fmt.Errorf("unmarshal op output: %w", err)
	}

	if opErr != nil {
		// The read itself succeeded; the server reported the error inside the
		// frame it handed back.
		timer.failWith(PhaseRead, PhaseServer)

		return opErr
	}

	// Consumed before the timer closes: the copy is work the caller waited on.
	if err := readResponse(response); err != nil {
		timer.fail(PhaseRead)
		return fmt.Errorf("unmarshal op response: %w", err)
	}

	timer.succeed(PhaseRead)

	return nil
}

// writeInputFrame encodes OpInput{operation: op, msg: <appendMsg>,
// frame_version: FrameVersion} straight
// into the lane's input region, behind the little-endian frame length the
// server reads first. appendMsg must append exactly msgSize bytes!!! The bytes
// match proto.Marshal, which omits zero-valued proto3 fields.
func writeInputFrame(
	region []byte,
	op shmpb.OperationType,
	msgSize int,
	appendMsg func(dst []byte) ([]byte, error),
) error {
	opVarint := uint64(op) //nolint:gosec // proto encodes int32 enums sign-extended

	var size int
	if opVarint != 0 {
		size += protowire.SizeTag(opInputOperationField) + protowire.SizeVarint(opVarint)
	}

	if msgSize > 0 {
		size += protowire.SizeTag(opInputMsgField) + protowire.SizeBytes(msgSize)
	}

	size += protowire.SizeTag(opInputFrameVersionField) + protowire.SizeVarint(uint64(FrameVersion))

	if size > len(region)-shm.FrameHeaderSize {
		return fmt.Errorf("%w: %d > %d bytes", errFrameTooLarge, size, len(region)-shm.FrameHeaderSize)
	}

	frameLen, err := mathutil.SafeIntToUint32(size)
	if err != nil {
		return fmt.Errorf("%w: %w", errFrameTooLarge, err)
	}

	// The capacity limit makes an append past the region reallocate instead of
	// overwriting it. The length check below rejects that case.
	frame := region[shm.FrameHeaderSize:shm.FrameHeaderSize:len(region)]
	if opVarint != 0 {
		frame = protowire.AppendTag(frame, opInputOperationField, protowire.VarintType)
		frame = protowire.AppendVarint(frame, opVarint)
	}

	if msgSize > 0 {
		frame = protowire.AppendTag(frame, opInputMsgField, protowire.BytesType)
		frame = protowire.AppendVarint(frame, mathutil.MustSafeIntToUint64(msgSize))

		frame, err = appendMsg(frame)
		if err != nil {
			return fmt.Errorf("marshal op input: %w", err)
		}
	}

	// proto.Marshal writes fields in number order, so frame_version goes last.
	frame = protowire.AppendTag(frame, opInputFrameVersionField, protowire.VarintType)
	frame = protowire.AppendVarint(frame, uint64(FrameVersion))

	if len(frame) != size {
		return fmt.Errorf("marshal op input: encoded %d bytes, expected %d", len(frame), size)
	}

	binary.LittleEndian.PutUint32(region[:shm.FrameHeaderSize], frameLen)

	return nil
}

// parseOpOutput parses an OpOutput frame in place. response aliases frame. A
// present error field is a server error even when empty, as a nil check on the
// unmarshaled message would report it. Repeated error fields merge, as
// proto.Unmarshal merges repeated occurrences of a message field.
func parseOpOutput(frame []byte) (response []byte, opErr *shmpb.OpError, err error) {
	for len(frame) > 0 {
		num, typ, n := protowire.ConsumeTag(frame)
		if n < 0 {
			return nil, nil, protowire.ParseError(n)
		}

		frame = frame[n:]

		switch {
		case num == opOutputResponseField && typ == protowire.BytesType:
			response, n = protowire.ConsumeBytes(frame)

		case num == opOutputErrorField && typ == protowire.BytesType:
			var errBz []byte
			if errBz, n = protowire.ConsumeBytes(frame); n >= 0 {
				if opErr == nil {
					opErr = new(shmpb.OpError)
				}
				if err := (proto.UnmarshalOptions{Merge: true}).Unmarshal(errBz, opErr); err != nil {
					return nil, nil, fmt.Errorf("unmarshal op error: %w", err)
				}
			}

		default:
			n = protowire.ConsumeFieldValue(num, typ, frame)
		}

		if n < 0 {
			return nil, nil, protowire.ParseError(n)
		}

		frame = frame[n:]
	}

	if opErr == nil {
		return response, nil, nil
	}

	return nil, opErr, nil
}

// claimLane will attempt to claim an idle lane from the lane set. It returns a
// claimedLane struct containing the shared memory region, lane index, and
// generation number if successful. If no idle lanes are available or if the
// lane set is closed, it returns an error.
//
// We perform a round-robin lane selection to ensure that all lanes are utilized
// evenly. We evaluate each lane's state and attempt to lock it for exclusive
// access. If the lane is idle, we claim it and return the claimedLane struct.
// If the lane is NOT idle or if we FAIL to lock it, then we consider this to be
// an invariant failure and we return an error:
//
// - GetMem(lane) can only fail if the lane-set is invalid/closed or the calculated
// lane index is invalid. Continuing would hide that fault and eventually report
// the misleading “no idle lanes available.”
//
// - TryLockLane(lane) returns (false, nil) when another local goroutine owns the
// lock—that is the normal case, and claimLane does continue. A non-nil error
// likewise means an invalid lane-set/index.
func (s *Service) claimLane() (*claimedLane, error) {
	if s.ls == nil || s.lanes == 0 {
		return nil, fmt.Errorf("lane set is closed")
	}

	start := s.nextLaneIndex()
	for offset := range s.lanes {
		lane := (start + offset) % s.lanes

		mem, err := s.ls.GetMem(lane)
		if err != nil {
			return nil, fmt.Errorf("get lane %d memory: %w", lane, err)
		}

		if mem.LoadState() != shm.StateIdle {
			continue
		}

		locked, err := s.ls.TryLockLane(lane)
		if err != nil {
			return nil, fmt.Errorf("try lock lane %d: %w", lane, err)
		}

		if !locked {
			continue
		}

		generation, claimed := mem.TryClaimIdle()
		if claimed {
			return &claimedLane{mem: mem, lane: lane, generation: generation}, nil
		}

		if err := s.ls.UnlockLane(lane); err != nil {
			return nil, fmt.Errorf("unlock lane %d: %w", lane, err)
		}
	}

	return nil, fmt.Errorf("no idle lanes available")
}

// nextLaneIndex returns the next round-robin lane index and keeps nextLane
// within the configured lane range.
func (s *Service) nextLaneIndex() int {
	// Use a CAS loop to ensure that nextLane is updated atomically and wraps around
	// the lane count.
	lanes := mathutil.MustSafeIntToUint64(s.lanes)
	for {
		current := s.nextLane.Load()
		next := (current + 1) % lanes

		if s.nextLane.CompareAndSwap(current, next) {
			return int(current) //nolint:gosec // lanes are capped to a small number
		}
	}
}

// waitForState blocks until mem reaches want, the lane exits, or ctx ends. It
// waits on the state futex in 1 ms slices so ctx is checked promptly.
func waitForState(ctx context.Context, mem *shm.Mem, want uint32) error {
	stateAddr := mem.StateAddr()

	for {
		state := mem.LoadState()
		switch state {
		case want:
			return nil

		case shm.StateExit:
			return fmt.Errorf("lane exited before operation completed")

		default:
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("timed out waiting for state=%d (current=%d): %w", want, state, err)
			}

			err := shm.FutexWaitTimeout(stateAddr, state, time.Millisecond)
			if errors.Is(err, syscall.ETIMEDOUT) {
				continue
			}
			if err != nil {
				return fmt.Errorf("futex wait: %w", err)
			}
		}
	}
}

// newPhaseTimer starts timing op. The admission wait is its first phase.
func (s *Service) newPhaseTimer(op string) *phaseTimer {
	now := time.Now()
	return &phaseTimer{svc: s, op: op, start: now, last: now}
}

// markAdmission closes the admission wait and starts the first real phase.
func (t *phaseTimer) markAdmission() {
	now := time.Now()
	if t.svc.observer == nil {
		t.last = now
		return
	}

	t.svc.observer.AdmissionWait(t.op, now.Sub(t.last))

	// Restart from after the callback, so the sink's own cost is billed to nobody
	// rather than to the phase that follows it.
	t.last = time.Now()
}

// mark closes the named phase and starts the next.
func (t *phaseTimer) mark(phase string) {
	now := time.Now()
	if t.svc.observer == nil {
		t.last = now
		return
	}

	t.svc.observer.Phase(t.op, phase, now.Sub(t.last))

	t.last = time.Now()
}

// succeed closes the final phase of an operation that produced a response.
func (t *phaseTimer) succeed(phase string) {
	t.finish(phase, OutcomeSuccess)
}

// fail records that the operation ended in phase, and reports the phase and
// operation durations too so an operation that gave up still contributes its
// time.
func (t *phaseTimer) fail(phase string) {
	t.failWith(phase, phase)
}

// failWith closes phase but blames cause, for an error that surfaces in one
// phase while originating in another.
func (t *phaseTimer) failWith(phase, cause string) {
	t.finish(phase, OutcomeFailure)

	if t.svc.observer != nil {
		t.svc.observer.Failure(t.op, cause)
	}
}

// failBeforeAdmission records an operation rejected before it could wait for a
// lane. There is no phase to close, but it still has to appear end to end or
// the failure count outruns the operation count.
func (t *phaseTimer) failBeforeAdmission() {
	if t.svc.observer == nil {
		return
	}

	t.svc.observer.Operation(t.op, OutcomeFailure, time.Since(t.start))
	t.svc.observer.Failure(t.op, PhaseAdmission)
}

// resume restarts the phase clock, for a caller that ran an observer callback
// of its own between two phases.
func (t *phaseTimer) resume() {
	t.last = time.Now()
}

// finish closes the last phase an operation reached and reports the operation
// end to end, which a scrape cannot reconstruct by summing phase quantiles.
func (t *phaseTimer) finish(phase, outcome string) {
	now := time.Now()
	if t.svc.observer == nil {
		return
	}

	t.svc.observer.Phase(t.op, phase, now.Sub(t.last))
	t.svc.observer.Operation(t.op, outcome, now.Sub(t.start))
}
