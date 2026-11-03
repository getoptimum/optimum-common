package shm

import "github.com/getoptimum/shm/pkg/shm"

// LaneSet defines the interface for a shared memory lane set, which is used to
// manage and access shared memory lanes.
//
// Note, this interface exists primarily for mocking purposes in tests.
type LaneSet interface {
	GetMem(lane int) (*shm.Mem, error)
	LockLane(lane int) error
	TryLockLane(lane int) (bool, error)
	UnlockLane(lane int) error
}
