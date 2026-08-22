package syncx_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

// Race-detector test: quiet without -race, reports the unsynchronised
// activeListeners read under -race when the fix is absent.
func TestBroadcastTry_ConcurrentWithRegistration(_ *testing.T) {
	b := syncx.NewBroadcaster[int]()

	const workers = 8
	var wg sync.WaitGroup

	for i := range workers {
		wg.Go(func() {
			key := "listener-" + strconv.Itoa(i)
			for range 200 {
				ch := b.RegisterBufferedListener(key, 1)
				select {
				case <-ch:
				default:
				}
				b.UnregisterListener(key)
			}
		})
	}

	for range workers {
		wg.Go(func() {
			for range 200 {
				b.BroadcastTry(1, nil)
			}
		})
	}

	wg.Wait()
}
