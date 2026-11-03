package shm

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	shmpb "github.com/getoptimum/optimum-common/pkg/rlncpb/shm"
	rlncpbtypes "github.com/getoptimum/optimum-common/pkg/rlncpb/types"
	libshm "github.com/getoptimum/shm/pkg/shm"
)

const testFrameVersion uint32 = 1

func TestNewWithLaneSet(t *testing.T) {
	tests := []struct {
		name  string
		lanes int
	}{
		{
			name:  "stores lane set and lane count",
			lanes: 3,
		},
		{
			name:  "allows zero lane count",
			lanes: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ls := newLaneSetStub()

			svc := NewWithLaneSet(ls, tc.lanes, testFrameVersion)

			require.NotNil(t, svc)
			require.Equal(t, ls, svc.ls)
			require.Equal(t, tc.lanes, svc.lanes)
			require.Equal(t, max(0, tc.lanes), cap(svc.slots))
		})
	}
}

func TestServiceClaimLane(t *testing.T) {
	t.Run("claims selected idle lane", func(t *testing.T) {
		idleMem := newTestMem(t)
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(_ int) (*libshm.Mem, error) {
				return idleMem, nil
			}
		})
		svc := NewWithLaneSet(ls, 2, testFrameVersion)
		claimed, err := svc.claimLane()
		require.NoError(t, err)
		require.Equal(t, 0, claimed.lane)
		require.True(t, claimed.mem.ReleaseClaim(claimed.generation))
	})

	t.Run("skips busy lane and claims next idle lane", func(t *testing.T) {
		busyMem := newTestMem(t)
		busyMem.StoreState(libshm.StateInputReady)
		idleMem := newTestMem(t)
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(lane int) (*libshm.Mem, error) {
				if lane == 0 {
					return busyMem, nil
				}
				return idleMem, nil
			}
		})
		svc := NewWithLaneSet(ls, 2, testFrameVersion)

		claimed, err := svc.claimLane()
		require.NoError(t, err)
		require.Equal(t, 1, claimed.lane)
		require.NoError(t, ls.UnlockLane(claimed.lane))
		require.True(t, claimed.mem.ReleaseClaim(claimed.generation))
	})
}

func TestServiceClaimLaneFailures(t *testing.T) {
	t.Run("rejects closed or empty lane set", func(t *testing.T) {
		for _, svc := range []*Service{
			NewWithLaneSet(nil, 1, testFrameVersion),
			NewWithLaneSet(newLaneSetStub(), 0, testFrameVersion),
		} {
			_, err := svc.claimLane()
			require.ErrorContains(t, err, "lane set is closed")
		}
	})

	t.Run("returns lane lookup error", func(t *testing.T) {
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(_ int) (*libshm.Mem, error) {
				return nil, errors.New("lookup failed")
			}
		})

		_, err := NewWithLaneSet(ls, 1, testFrameVersion).claimLane()
		require.ErrorContains(t, err, "get lane 0 memory: lookup failed")
	})

	t.Run("returns lock error", func(t *testing.T) {
		mem := newTestMem(t)
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(_ int) (*libshm.Mem, error) { return mem, nil }
			ls.tryLockLaneFn = func(_ int) (bool, error) { return false, errors.New("lock failed") }
		})

		_, err := NewWithLaneSet(ls, 1, testFrameVersion).claimLane()
		require.ErrorContains(t, err, "try lock lane 0: lock failed")
		require.Equal(t, 1, ls.tryLockCalled)
	})

	t.Run("skips locally locked lane", func(t *testing.T) {
		mem := newTestMem(t)
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(_ int) (*libshm.Mem, error) { return mem, nil }
			ls.tryLockLaneFn = func(_ int) (bool, error) { return false, nil }
		})

		_, err := NewWithLaneSet(ls, 1, testFrameVersion).claimLane()
		require.ErrorContains(t, err, "no idle lanes available")
		require.Equal(t, 1, ls.tryLockCalled)
		require.Zero(t, ls.unlockCalled)
	})

	t.Run("unlocks and continues when claim loses race", func(t *testing.T) {
		first := newTestMem(t)
		second := newTestMem(t)
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(lane int) (*libshm.Mem, error) {
				if lane == 0 {
					return first, nil
				}
				return second, nil
			}
			ls.tryLockLaneFn = func(lane int) (bool, error) {
				if lane == 0 {
					first.StoreState(libshm.StateInputReady)
				}
				return true, nil
			}
		})

		claimed, err := NewWithLaneSet(ls, 2, testFrameVersion).claimLane()
		require.NoError(t, err)
		require.Equal(t, 1, claimed.lane)
		require.Equal(t, 1, ls.unlockCalled)
		require.NoError(t, ls.UnlockLane(claimed.lane))
		require.True(t, claimed.mem.ReleaseClaim(claimed.generation))
	})

	t.Run("returns unlock error after lost claim", func(t *testing.T) {
		mem := newTestMem(t)
		ls := newLaneSetStub(func(ls *laneSetStub) {
			ls.getMemFn = func(_ int) (*libshm.Mem, error) { return mem, nil }
			ls.tryLockLaneFn = func(_ int) (bool, error) {
				mem.StoreState(libshm.StateInputReady)
				return true, nil
			}
			ls.unlockLaneFn = func(_ int) error { return errors.New("unlock failed") }
		})

		_, err := NewWithLaneSet(ls, 1, testFrameVersion).claimLane()
		require.ErrorContains(t, err, "unlock lane 0: unlock failed")
	})
}

func TestServiceNextLaneIndexWraps(t *testing.T) {
	svc := NewWithLaneSet(newLaneSetStub(), 2, testFrameVersion)

	require.Equal(t, 0, svc.nextLaneIndex())
	require.Equal(t, 1, svc.nextLaneIndex())
	require.Equal(t, 0, svc.nextLaneIndex())
}

// TestServiceExecuteOpWaitsForLocalCapacity covers the admission gate's central
// behavior: a caller that finds every slot taken waits and then resumes, rather
// than entering claimLane where the only lane being locally locked would surface
// as a premature "no idle lanes available" failure.
//
// The stub fails inside GetMem, which keeps both operations off the futex path
// while still exercising the gate that wraps them.
func TestServiceExecuteOpWaitsForLocalCapacity(t *testing.T) {
	var (
		mu          sync.Mutex
		getMemCalls int
	)

	entered := make(chan struct{}, 1)
	release := make(chan struct{})

	ls := newLaneSetStub(func(ls *laneSetStub) {
		ls.getMemFn = func(_ int) (*libshm.Mem, error) {
			mu.Lock()
			getMemCalls++
			first := getMemCalls == 1
			mu.Unlock()

			// Hold the only slot until the test releases it.
			if first {
				entered <- struct{}{}
				<-release
			}

			return nil, errors.New("lane unavailable")
		}
	})

	laneSetCalls := func() int {
		mu.Lock()
		defer mu.Unlock()

		return getMemCalls
	}

	svc := NewWithLaneSet(ls, 1, testFrameVersion)

	first := make(chan error, 1)
	go func() {
		_, err := svc.ExecuteOp(shmpb.OperationType_OPERATION_TYPE_DECODE, nil)
		first <- err
	}()

	select {
	case <-entered:

	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("first operation never reached the lane set")
	}

	second := make(chan error, 1)
	go func() {
		_, err := svc.ExecuteOp(shmpb.OperationType_OPERATION_TYPE_DECODE, nil)
		second <- err
	}()

	select {
	case err := <-second:
		close(release)
		t.Fatalf("second operation did not wait for local capacity: %v", err)

	case <-time.After(100 * time.Millisecond):
	}

	require.Equal(t, 1, laneSetCalls(), "a waiting caller must not enter the lane set")

	close(release)

	require.Error(t, <-first)

	select {
	case err := <-second:
		require.Error(t, err)
		require.NotContains(
			t,
			err.Error(),
			"no idle lanes available",
			"waiting for capacity must not degrade into a lane exhaustion failure",
		)

	case <-time.After(5 * time.Second):
		t.Fatal("second operation did not resume after the first released its slot")
	}

	require.Equal(t, 2, laneSetCalls(), "the waiting caller must proceed once a slot frees")
}

// TestServiceExecuteOpBoundsConcurrencyToLaneCount asserts the gate admits at
// most one caller per physical lane, which is what keeps local contention out of
// claimLane entirely.
func TestServiceExecuteOpBoundsConcurrencyToLaneCount(t *testing.T) {
	const (
		lanes   = 3
		callers = 24
	)

	var (
		mu           sync.Mutex
		inFlight     int
		peakInFlight int
	)

	ls := newLaneSetStub(func(ls *laneSetStub) {
		ls.getMemFn = func(_ int) (*libshm.Mem, error) {
			mu.Lock()
			inFlight++
			peakInFlight = max(peakInFlight, inFlight)
			mu.Unlock()

			time.Sleep(time.Millisecond)

			mu.Lock()
			inFlight--
			mu.Unlock()

			return nil, errors.New("lane unavailable")
		}
	})

	svc := NewWithLaneSet(ls, lanes, testFrameVersion)

	errs := make(chan error, callers)

	var wg sync.WaitGroup
	for range callers {
		wg.Go(func() {
			_, err := svc.ExecuteOp(shmpb.OperationType_OPERATION_TYPE_DECODE, nil)
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	// Every caller reaches the lane set; the stub then fails each of them.
	for err := range errs {
		require.Error(t, err)
	}

	mu.Lock()
	defer mu.Unlock()

	require.Positive(t, peakInFlight)
	require.LessOrEqual(t, peakInFlight, lanes, "no more callers than lanes may enter the lane set at once")
}

func TestServiceExecuteOp(t *testing.T) {
	tests := []struct {
		name            string
		configureLS     func(*laneSetStub, *libshm.Mem)
		worker          func(*testing.T, *libshm.Mem, shmpb.OperationType, []byte) <-chan opWorkerResult
		assertWorker    func(*testing.T, opWorkerResult)
		op              shmpb.OperationType
		payload         []byte
		wantResp        []byte
		wantErrSubstr   string
		wantLockCalls   int
		wantUnlockCalls int
	}{
		{
			name: "returns claim lane error",
			configureLS: func(ls *laneSetStub, _ *libshm.Mem) {
				ls.getMemFn = func(_ int) (*libshm.Mem, error) {
					return nil, errors.New("no lanes")
				}
			},
			op:              shmpb.OperationType_OPERATION_TYPE_LEGACY_ENCODE_INTO_SHARDS,
			payload:         []byte("req"),
			wantErrSubstr:   "get lane 0 memory: no lanes",
			wantLockCalls:   0,
			wantUnlockCalls: 0,
		},
		{
			name: "returns response bytes from worker",
			configureLS: func(ls *laneSetStub, mem *libshm.Mem) {
				ls.getMemFn = func(_ int) (*libshm.Mem, error) {
					return mem, nil
				}
			},
			worker: startWorkerWithOpOutput(
				func(_ *shmpb.OpInput, _ shmpb.OperationType, _ []byte) *shmpb.OpOutput {
					return &shmpb.OpOutput{Response: []byte("ok")}
				},
			),
			assertWorker: func(t *testing.T, result opWorkerResult) {
				t.Helper()
				require.NotNil(t, result.request)
				require.Equal(t, result.op, result.request.GetOperation())
				require.Equal(t, result.payload, result.request.GetMsg())
			},
			op:              shmpb.OperationType_OPERATION_TYPE_LEGACY_DECODE_FROM_SHARDS,
			payload:         []byte("hello"),
			wantResp:        []byte("ok"),
			wantLockCalls:   1,
			wantUnlockCalls: 1,
		},
		{
			name: "returns operation error from worker",
			configureLS: func(ls *laneSetStub, mem *libshm.Mem) {
				ls.getMemFn = func(_ int) (*libshm.Mem, error) {
					return mem, nil
				}
			},
			worker: startWorkerWithOpOutput(func(_ *shmpb.OpInput, _ shmpb.OperationType, _ []byte) *shmpb.OpOutput {
				return &shmpb.OpOutput{
					Error: &shmpb.OpError{Code: shmpb.OpErrorCode_OP_ERROR_CODE_UNMARSHAL_FAILED, Err: "decode failed"},
				}
			}),
			op:              shmpb.OperationType_OPERATION_TYPE_LEGACY_RECODE_FROM_SHARDS,
			payload:         []byte("payload"),
			wantErrSubstr:   "decode failed",
			wantLockCalls:   1,
			wantUnlockCalls: 1,
		},
		{
			name: "returns unmarshal error on malformed output",
			configureLS: func(ls *laneSetStub, mem *libshm.Mem) {
				ls.getMemFn = func(_ int) (*libshm.Mem, error) {
					return mem, nil
				}
			},
			worker:          startWorkerWithRawOutput([]byte("not-a-protobuf")),
			op:              shmpb.OperationType_OPERATION_TYPE_LEGACY_ENCODE_INTO_SHARDS,
			payload:         []byte("x"),
			wantErrSubstr:   "unmarshal op output",
			wantLockCalls:   1,
			wantUnlockCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mem := newTestMem(t)
			ls := newLaneSetStub(func(ls *laneSetStub) {
				ls.getRandMemFn = func() (*libshm.Mem, int, error) {
					return mem, 0, nil
				}
			})
			if tc.configureLS != nil {
				tc.configureLS(ls, mem)
			}

			svc := NewWithLaneSet(ls, 1, testFrameVersion)

			var workerDone <-chan opWorkerResult
			if tc.worker != nil {
				workerDone = tc.worker(t, mem, tc.op, tc.payload)
			}

			gotResp, err := svc.ExecuteOp(tc.op, tc.payload)
			if tc.wantErrSubstr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErrSubstr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantResp, gotResp)
			}

			if workerDone != nil {
				result := <-workerDone
				require.NoError(t, result.err)
				if tc.assertWorker != nil {
					tc.assertWorker(t, result)
				}
			}

			if tc.worker != nil {
				require.Equal(t, libshm.StateIdle, mem.LoadState())
				require.Equal(t, byte(0), mem.Input()[0])
				require.Equal(t, byte(0), mem.Output()[0])
			}
			require.Equal(t, tc.wantLockCalls, ls.tryLockCalled)
			require.Equal(t, tc.wantUnlockCalls, ls.unlockCalled)
		})
	}
}

// startWorkerWithOpOutput returns a worker that answers one request with the
// marshaled OpOutput builder returns.
func startWorkerWithOpOutput(
	builder func(*shmpb.OpInput, shmpb.OperationType, []byte) *shmpb.OpOutput,
) func(*testing.T, *libshm.Mem, shmpb.OperationType, []byte) <-chan opWorkerResult {
	return startWorkerWithResponseBytes(
		func(req *shmpb.OpInput, op shmpb.OperationType, payload []byte) ([]byte, error) {
			response := builder(req, op, payload)

			responseBz, err := proto.Marshal(response)
			if err != nil {
				return nil, fmt.Errorf("marshal output: %w", err)
			}

			return responseBz, nil
		},
	)
}

// startWorkerWithRawOutput returns a worker that answers one request with raw
// output frame bytes, valid or not.
func startWorkerWithRawOutput(
	raw []byte,
) func(*testing.T, *libshm.Mem, shmpb.OperationType, []byte) <-chan opWorkerResult {
	return startWorkerWithResponseBytes(func(_ *shmpb.OpInput, _ shmpb.OperationType, _ []byte) ([]byte, error) {
		return bytes.Clone(raw), nil
	})
}

// startWorkerWithResponseBytes returns a worker that plays rlnc-server for one
// request on mem. It accepts the input frame, decodes it as OpInput, writes the
// built response frame, and publishes it. The result channel reports the
// decoded request and any protocol step that failed.
func startWorkerWithResponseBytes(
	responseBuilder func(*shmpb.OpInput, shmpb.OperationType, []byte) ([]byte, error),
) func(*testing.T, *libshm.Mem, shmpb.OperationType, []byte) <-chan opWorkerResult {
	return func(t *testing.T, mem *libshm.Mem, op shmpb.OperationType, payload []byte) <-chan opWorkerResult {
		t.Helper()

		done := make(chan opWorkerResult, 1)
		go func() {
			result := opWorkerResult{op: op, payload: bytes.Clone(payload)}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if err := waitForState(ctx, mem, libshm.StateInputReady); err != nil {
				result.err = fmt.Errorf("wait for input ready: %w", err)
				done <- result
				return
			}
			generation, accepted := mem.TryAcceptRequest()
			if !accepted {
				result.err = errors.New("accept request")
				done <- result
				return
			}

			requestBz, err := mem.ReadInputFrame()
			if err != nil {
				result.err = fmt.Errorf("read input frame: %w", err)
				done <- result
				return
			}

			request := new(shmpb.OpInput)
			if err := proto.Unmarshal(requestBz, request); err != nil {
				result.err = fmt.Errorf("unmarshal input: %w", err)
				done <- result
				return
			}
			result.request = request

			responseBz, err := responseBuilder(request, op, payload)
			if err != nil {
				result.err = err
				done <- result
				return
			}

			if err := mem.WriteOutputFrame(responseBz); err != nil {
				result.err = fmt.Errorf("write output frame: %w", err)
				done <- result
				return
			}

			mem.StoreOutputReadyAt(time.Now().UnixNano())
			if !mem.PublishOutput(generation) {
				result.err = errors.New("publish output")
				done <- result
				return
			}
			if err := libshm.FutexWake(mem.StateAddr(), 1); err != nil {
				result.err = fmt.Errorf("futex wake output ready: %w", err)
				done <- result
				return
			}

			done <- result
		}()

		return done
	}
}

// opWorkerResult is what a test worker observed while serving one request.
type opWorkerResult struct {
	err     error
	request *shmpb.OpInput
	op      shmpb.OperationType
	payload []byte
}

// newTestMem maps an idle lane backed by a temporary file, unmapped at cleanup.
func newTestMem(t *testing.T) *libshm.Mem {
	t.Helper()

	path := filepath.Join(t.TempDir(), "lane.mem")
	mem, err := libshm.CreateAt(path)
	require.NoError(t, err)
	mem.StoreState(libshm.StateIdle)

	t.Cleanup(func() {
		require.NoError(t, mem.Unmap())
	})

	return mem
}

// laneSetStub is a LaneSet whose behavior each test overrides per method. It
// counts lock calls so tests can assert lane release.
type laneSetStub struct {
	getMemFn      func(lane int) (*libshm.Mem, error)
	getRandMemFn  func() (*libshm.Mem, int, error)
	lockLaneFn    func(lane int) error
	tryLockLaneFn func(lane int) (bool, error)
	unlockLaneFn  func(lane int) error

	mu            sync.Mutex
	lockCalled    int
	tryLockCalled int
	unlockCalled  int
}

// newLaneSetStub returns a stub whose locks always succeed, adjusted by options.
func newLaneSetStub(options ...func(*laneSetStub)) *laneSetStub {
	ls := &laneSetStub{
		getRandMemFn: func() (*libshm.Mem, int, error) {
			return nil, 0, nil
		},
		lockLaneFn: func(_ int) error {
			return nil
		},
		tryLockLaneFn: func(_ int) (bool, error) {
			return true, nil
		},
		unlockLaneFn: func(_ int) error {
			return nil
		},
	}

	for _, opt := range options {
		opt(ls)
	}

	return ls
}

// GetMem returns getMemFn's lane, falling back to getRandMemFn.
func (l *laneSetStub) GetMem(lane int) (*libshm.Mem, error) {
	if l.getMemFn != nil {
		return l.getMemFn(lane)
	}
	mem, _, err := l.getRandMemFn()
	return mem, err
}

// GetRandMem delegates to getRandMemFn.
func (l *laneSetStub) GetRandMem() (*libshm.Mem, int, error) {
	return l.getRandMemFn()
}

// LockLane counts the call and delegates to lockLaneFn.
func (l *laneSetStub) LockLane(lane int) error {
	l.mu.Lock()
	l.lockCalled++
	l.mu.Unlock()

	return l.lockLaneFn(lane)
}

// TryLockLane counts the call and delegates to tryLockLaneFn.
func (l *laneSetStub) TryLockLane(lane int) (bool, error) {
	l.mu.Lock()
	l.tryLockCalled++
	l.mu.Unlock()

	return l.tryLockLaneFn(lane)
}

// UnlockLane counts the call and delegates to unlockLaneFn.
func (l *laneSetStub) UnlockLane(lane int) error {
	l.mu.Lock()
	l.unlockCalled++
	l.mu.Unlock()

	return l.unlockLaneFn(lane)
}

func TestPhaseTimerReportsOperationEndToEnd(t *testing.T) {
	obs := &observerStub{}
	svc := NewWithLaneSet(newLaneSetStub(), 1, testFrameVersion)
	svc.SetObserver(obs)

	timer := svc.newPhaseTimer("TEST_OP")
	timer.markAdmission()
	timer.mark(PhaseClaim)
	timer.succeed(PhaseRead)

	require.Len(t, obs.admissions, 1)
	require.Len(t, obs.phases, 2)
	require.Empty(t, obs.failures)

	require.Len(t, obs.operations, 1)
	require.Equal(t, "TEST_OP", obs.operations[0].op)
	require.Equal(t, OutcomeSuccess, obs.operations[0].outcome)

	// The operation covers the admission wait and every phase, which is exactly
	// what summing phase histograms cannot recover.
	var phaseSum time.Duration
	for _, p := range obs.phases {
		phaseSum += p.d
	}

	require.GreaterOrEqual(t, obs.operations[0].d, obs.admissions[0].d+phaseSum)
}

func TestPhaseTimerFailReportsFailedOperation(t *testing.T) {
	obs := &observerStub{}
	svc := NewWithLaneSet(newLaneSetStub(), 1, testFrameVersion)
	svc.SetObserver(obs)

	timer := svc.newPhaseTimer("TEST_OP")
	timer.markAdmission()
	timer.fail(PhaseWait)

	require.Len(t, obs.phases, 1)
	require.Equal(t, PhaseWait, obs.phases[0].phase)

	require.Len(t, obs.failures, 1)
	require.Equal(t, PhaseWait, obs.failures[0].phase)

	require.Len(t, obs.operations, 1)
	require.Equal(t, OutcomeFailure, obs.operations[0].outcome)
}

func TestPhaseTimerExcludesObserverCostFromNextPhase(t *testing.T) {
	const observerCost = 50 * time.Millisecond

	obs := &observerStub{cost: observerCost}
	svc := NewWithLaneSet(newLaneSetStub(), 1, testFrameVersion)
	svc.SetObserver(obs)

	timer := svc.newPhaseTimer("TEST_OP")
	timer.markAdmission()
	timer.mark(PhaseClaim)
	timer.succeed(PhaseRead)

	// Each phase starts after the previous callback returned, so a slow sink
	// cannot show up as time the operation spent working.
	for _, p := range obs.phases {
		require.Less(t, p.d, observerCost, "phase %s absorbed observer cost", p.phase)
	}
}

type (
	// observedPhase is one phase, admission, or failure report.
	observedPhase struct {
		phase string
		d     time.Duration
	}

	// observedOp is one end-to-end operation report.
	observedOp struct {
		op      string
		outcome string
		d       time.Duration
	}

	// observerStub records observations and can simulate a slow metric sink.
	observerStub struct {
		mu   sync.Mutex
		cost time.Duration

		admissions []observedPhase
		phases     []observedPhase
		operations []observedOp
		failures   []observedPhase
	}
)

// record applies fn under the lock, then sleeps for the simulated sink cost.
func (o *observerStub) record(fn func()) {
	o.mu.Lock()
	defer o.mu.Unlock()

	fn()

	time.Sleep(o.cost)
}

// AdmissionWait records an admission wait.
func (o *observerStub) AdmissionWait(_ string, d time.Duration) {
	o.record(func() {
		o.admissions = append(o.admissions, observedPhase{phase: PhaseAdmission, d: d})
	})
}

// Phase records a completed phase.
func (o *observerStub) Phase(_, phase string, d time.Duration) {
	o.record(func() {
		o.phases = append(o.phases, observedPhase{phase: phase, d: d})
	})
}

// Operation records an end-to-end operation.
func (o *observerStub) Operation(op, outcome string, d time.Duration) {
	o.record(func() {
		o.operations = append(o.operations, observedOp{op: op, outcome: outcome, d: d})
	})
}

// Failure records the phase an operation failed in.
func (o *observerStub) Failure(_, phase string) {
	o.record(func() {
		o.failures = append(o.failures, observedPhase{phase: phase})
	})
}

// LaneClaimed ignores lane occupancy, which these tests do not assert.
func (o *observerStub) LaneClaimed() {}

// LaneReleased ignores lane occupancy, which these tests do not assert.
func (o *observerStub) LaneReleased() {}

func TestServiceExecuteOpReportsRejectionBeforeAdmission(t *testing.T) {
	obs := &observerStub{}
	svc := NewWithLaneSet(newLaneSetStub(), 0, testFrameVersion)
	svc.SetObserver(obs)

	_, err := svc.ExecuteOp(shmpb.OperationType_OPERATION_TYPE_DECODE, nil)
	require.ErrorContains(t, err, "lane set is closed")

	// Every failure must have an end-to-end sample behind it, or the failure
	// count outruns the operation count for a closed service.
	require.Len(t, obs.failures, 1)
	require.Equal(t, PhaseAdmission, obs.failures[0].phase)

	require.Len(t, obs.operations, 1)
	require.Equal(t, OutcomeFailure, obs.operations[0].outcome)

	// It never got as far as waiting for or holding a lane.
	require.Empty(t, obs.admissions)
	require.Empty(t, obs.phases)
}

func TestPhaseTimerFailWithBlamesCauseNotPhase(t *testing.T) {
	obs := &observerStub{}
	svc := NewWithLaneSet(newLaneSetStub(), 1, testFrameVersion)
	svc.SetObserver(obs)

	timer := svc.newPhaseTimer("TEST_OP")
	timer.markAdmission()
	timer.failWith(PhaseRead, PhaseServer)

	// The read is still timed as a read, but a server-side error is not counted
	// against the client's read.
	require.Len(t, obs.phases, 1)
	require.Equal(t, PhaseRead, obs.phases[0].phase)

	require.Len(t, obs.failures, 1)
	require.Equal(t, PhaseServer, obs.failures[0].phase)
}

func TestPhaseTimerResumeExcludesCallerObserverCost(t *testing.T) {
	const observerCost = 50 * time.Millisecond

	obs := &observerStub{cost: observerCost}
	svc := NewWithLaneSet(newLaneSetStub(), 1, testFrameVersion)
	svc.SetObserver(obs)

	timer := svc.newPhaseTimer("TEST_OP")
	timer.markAdmission()
	timer.mark(PhaseClaim)

	// Mirrors ExecuteOp's LaneClaimed call between two phases.
	obs.LaneClaimed()
	timer.resume()

	timer.succeed(PhaseSubmit)

	require.Len(t, obs.phases, 2)
	require.Less(t, obs.phases[1].d, observerCost)
}

func TestServiceExecuteMessage(t *testing.T) {
	req := &rlncpbtypes.ShardRequest{Data: []byte("payload"), MaxShardSize: 4}
	reqBz, err := proto.Marshal(req)
	require.NoError(t, err)

	tests := []struct {
		name     string
		output   *shmpb.OpOutput
		wantResp *rlncpbtypes.ShardResponse
		wantErr  *shmpb.OpError
	}{
		{
			name: "decodes response",
			output: &shmpb.OpOutput{
				Response: mustMarshal(t, &rlncpbtypes.ShardResponse{Shards: [][]byte{[]byte("payl"), []byte("oad")}}),
			},
			wantResp: &rlncpbtypes.ShardResponse{Shards: [][]byte{[]byte("payl"), []byte("oad")}},
		},
		{
			name: "returns operation error",
			output: &shmpb.OpOutput{
				Error: &shmpb.OpError{Code: shmpb.OpErrorCode_OP_ERROR_CODE_INCONSISTENT_ROW, Err: "inconsistent"},
			},
			wantErr: &shmpb.OpError{Code: shmpb.OpErrorCode_OP_ERROR_CODE_INCONSISTENT_ROW, Err: "inconsistent"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mem := newTestMem(t)
			ls := newLaneSetStub(func(ls *laneSetStub) {
				ls.getMemFn = func(_ int) (*libshm.Mem, error) { return mem, nil }
			})
			svc := NewWithLaneSet(ls, 1, testFrameVersion)

			op := shmpb.OperationType_OPERATION_TYPE_SHARD
			workerDone := startWorkerWithOpOutput(
				func(_ *shmpb.OpInput, _ shmpb.OperationType, _ []byte) *shmpb.OpOutput { return tc.output },
			)(t, mem, op, reqBz)

			var resp rlncpbtypes.ShardResponse
			err := svc.ExecuteMessage(op, req, &resp)

			result := <-workerDone
			require.NoError(t, result.err)
			require.Equal(t, op, result.request.GetOperation())
			require.Equal(t, reqBz, result.request.GetMsg())

			if tc.wantErr != nil {
				opErr, ok := errors.AsType[*shmpb.OpError](err)
				require.True(t, ok, "want *OpError, got %v", err)
				require.True(t, proto.Equal(tc.wantErr, opErr))
			} else {
				require.NoError(t, err)
				require.True(t, proto.Equal(tc.wantResp, &resp))
			}

			require.Equal(t, libshm.StateIdle, mem.LoadState())
		})
	}
}

func TestServiceVerifyServer(t *testing.T) {
	tests := []struct {
		name        string
		output      *shmpb.OpOutput
		wantVersion string
		wantErr     error
	}{
		{
			name: "compatible server",
			output: &shmpb.OpOutput{
				Response: mustMarshal(t, &shmpb.ServerInfoResponse{
					FrameVersion:  testFrameVersion,
					ServerVersion: "v0.12.0",
				}),
			},
			wantVersion: "v0.12.0",
		},
		{
			name: "server predates frame versioning",
			output: &shmpb.OpOutput{
				Error: &shmpb.OpError{Code: shmpb.OpErrorCode_OP_ERROR_CODE_UNSUPPORTED_OPERATION, Err: "unsupported"},
			},
			wantErr: ErrIncompatibleServer,
		},
		{
			name: "different frame version",
			output: &shmpb.OpOutput{
				Response: mustMarshal(t, &shmpb.ServerInfoResponse{
					FrameVersion:  testFrameVersion + 1,
					ServerVersion: "v9.0.0",
				}),
			},
			wantVersion: "v9.0.0",
			wantErr:     ErrIncompatibleServer,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mem := newTestMem(t)
			ls := newLaneSetStub(func(ls *laneSetStub) {
				ls.getMemFn = func(_ int) (*libshm.Mem, error) { return mem, nil }
			})
			svc := NewWithLaneSet(ls, 1, testFrameVersion)

			op := shmpb.OperationType_OPERATION_TYPE_SERVER_INFO
			workerDone := startWorkerWithOpOutput(
				func(_ *shmpb.OpInput, _ shmpb.OperationType, _ []byte) *shmpb.OpOutput { return tc.output },
			)(t, mem, op, nil)

			version, err := svc.VerifyServer()

			result := <-workerDone
			require.NoError(t, result.err)
			require.Equal(t, op, result.request.GetOperation())
			require.Equal(t, testFrameVersion, result.request.GetFrameVersion())
			require.Equal(t, tc.wantVersion, version)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWriteInputFrameMatchesProtoMarshal(t *testing.T) {
	tests := []struct {
		name         string
		op           shmpb.OperationType
		frameVersion uint32
		req          proto.Message
	}{
		{
			name:         "decode request with symbols",
			op:           shmpb.OperationType_OPERATION_TYPE_DECODE,
			frameVersion: testFrameVersion,
			req: &rlncpbtypes.DecodeRequest{
				ReceivedSymbol: &rlncpbtypes.Symbol{Coefficients: []byte{1, 2}, Data: []byte("data")},
				LocalSymbols: rlncpbtypes.NewSymbolSet(
					&rlncpbtypes.Symbol{Coefficients: []byte{1, 0}, Data: []byte("a")},
				),
			},
		},
		{
			name:         "empty request",
			op:           shmpb.OperationType_OPERATION_TYPE_SHARD,
			frameVersion: testFrameVersion,
			req:          &rlncpbtypes.ShardRequest{},
		},
		{
			name:         "zero operation",
			op:           shmpb.OperationType_OPERATION_TYPE_UNSPECIFIED,
			frameVersion: testFrameVersion,
			req:          &rlncpbtypes.ShardRequest{Data: bytes.Repeat([]byte{7}, 300)},
		},
		{
			name:         "zero frame version",
			op:           shmpb.OperationType_OPERATION_TYPE_SHARD,
			frameVersion: 0,
			req:          &rlncpbtypes.ShardRequest{Data: []byte("payload")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			region := make([]byte, 1<<12)
			err := writeInputFrame(region, tc.op, tc.frameVersion, proto.Size(tc.req), func(dst []byte) ([]byte, error) {
				return proto.MarshalOptions{}.MarshalAppend(dst, tc.req)
			})
			require.NoError(t, err)

			want, err := proto.Marshal(&shmpb.OpInput{
				Operation:    tc.op,
				Msg:          mustMarshal(t, tc.req),
				FrameVersion: tc.frameVersion,
			})
			require.NoError(t, err)

			size := binary.LittleEndian.Uint32(region[:libshm.FrameHeaderSize])
			require.Equal(t, want, region[libshm.FrameHeaderSize:libshm.FrameHeaderSize+int(size)])
		})
	}
}

func TestWriteInputFrameRejectsInvalidRequests(t *testing.T) {
	appendN := func(n int) func([]byte) ([]byte, error) {
		return func(dst []byte) ([]byte, error) { return append(dst, make([]byte, n)...), nil }
	}

	t.Run("oversized request", func(t *testing.T) {
		region := make([]byte, 16)
		err := writeInputFrame(region, shmpb.OperationType_OPERATION_TYPE_DECODE, testFrameVersion, 32, appendN(32))
		require.ErrorIs(t, err, errFrameTooLarge)
		require.Equal(t, make([]byte, len(region)), region, "a rejected frame must not touch the region")
	})

	t.Run("marshal size mismatch", func(t *testing.T) {
		region := make([]byte, 64)
		err := writeInputFrame(region, shmpb.OperationType_OPERATION_TYPE_DECODE, testFrameVersion, 8, appendN(9))
		require.ErrorContains(t, err, "encoded 15 bytes, expected 14")
		require.Equal(t, make([]byte, libshm.FrameHeaderSize), region[:libshm.FrameHeaderSize],
			"a rejected frame must not publish a length")
	})

	t.Run("marshal error", func(t *testing.T) {
		region := make([]byte, 64)
		err := writeInputFrame(region, shmpb.OperationType_OPERATION_TYPE_DECODE, testFrameVersion, 1, func([]byte) ([]byte, error) {
			return nil, errors.New("boom")
		})
		require.ErrorContains(t, err, "marshal op input: boom")
	})
}

func TestParseOpOutputMatchesProtoUnmarshal(t *testing.T) {
	withUnknownField := func(frame []byte) []byte {
		frame = protowire.AppendTag(frame, 99, protowire.VarintType)
		return protowire.AppendVarint(frame, 42)
	}

	tests := []struct {
		name  string
		frame []byte
	}{
		{name: "empty frame", frame: nil},
		{name: "response", frame: mustMarshal(t, &shmpb.OpOutput{Response: []byte("ok")})},
		{name: "error", frame: mustMarshal(t, &shmpb.OpOutput{
			Error: &shmpb.OpError{Code: shmpb.OpErrorCode_OP_ERROR_CODE_PANIC_RECOVERED, Err: "panic"},
		})},
		{name: "empty error is still an error", frame: mustMarshal(t, &shmpb.OpOutput{Error: &shmpb.OpError{}})},
		{name: "error takes precedence over response", frame: mustMarshal(t, &shmpb.OpOutput{
			Response: []byte("ignored"),
			Error:    &shmpb.OpError{Err: "failed"},
		})},
		{
			name:  "unknown fields are skipped",
			frame: withUnknownField(mustMarshal(t, &shmpb.OpOutput{Response: []byte("ok")})),
		},
		{
			// Concatenated messages hold the error field twice.
			name: "repeated errors merge",
			frame: append(
				mustMarshal(t, &shmpb.OpOutput{Error: &shmpb.OpError{Code: shmpb.OpErrorCode_OP_ERROR_CODE_PANIC_RECOVERED}}),
				mustMarshal(t, &shmpb.OpOutput{Error: &shmpb.OpError{Err: "panic"}})...,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var want shmpb.OpOutput
			require.NoError(t, proto.Unmarshal(tc.frame, &want))

			response, opErr, err := parseOpOutput(tc.frame)
			require.NoError(t, err)

			if want.GetError() != nil {
				require.NotNil(t, opErr)
				require.True(t, proto.Equal(want.GetError(), opErr))
				require.Nil(t, response)
				return
			}

			require.Nil(t, opErr)
			require.True(t, bytes.Equal(want.GetResponse(), response))
		})
	}

	for name, frame := range map[string][]byte{
		"malformed":          []byte("not-a-protobuf"),
		"truncated response": mustMarshal(t, &shmpb.OpOutput{Response: []byte("ok")})[:3],
		"malformed error":    protowire.AppendBytes(protowire.AppendTag(nil, opOutputErrorField, protowire.BytesType), []byte{0xff}),
		"malformed earlier error": append(
			protowire.AppendBytes(protowire.AppendTag(nil, opOutputErrorField, protowire.BytesType), []byte{0xff}),
			mustMarshal(t, &shmpb.OpOutput{Error: &shmpb.OpError{Err: "panic"}})...,
		),
	} {
		t.Run(name, func(t *testing.T) {
			response, opErr, err := parseOpOutput(frame)
			require.Error(t, err)
			require.Nil(t, response)
			require.Nil(t, opErr)
		})
	}
}

// The hand-written frame codec must track the generated OpInput and OpOutput
// schemas.
func TestOpFrameFieldNumbersMatchSchema(t *testing.T) {
	input := (&shmpb.OpInput{}).ProtoReflect().Descriptor().Fields()
	require.Equal(t, input.ByName("operation").Number(), opInputOperationField)
	require.Equal(t, input.ByName("msg").Number(), opInputMsgField)
	require.Equal(t, input.ByName("frame_version").Number(), opInputFrameVersionField)

	output := (&shmpb.OpOutput{}).ProtoReflect().Descriptor().Fields()
	require.Equal(t, output.ByName("response").Number(), opOutputResponseField)
	require.Equal(t, output.ByName("error").Number(), opOutputErrorField)
}

// mustMarshal marshals msg, failing the test on error.
func mustMarshal(t *testing.T, msg proto.Message) []byte {
	t.Helper()

	bz, err := proto.Marshal(msg)
	require.NoError(t, err)

	return bz
}
