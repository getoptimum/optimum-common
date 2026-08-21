package syncx_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/syncx"
)

// TestBroadcastTry_ConcurrentWithRegistration exercises BroadcastTry against
// listeners being registered and unregistered at the same time. It is a race
// detector test: it passes either way without -race, and reports the
// unsynchronised read of activeListeners without the fix when -race is on.
func TestBroadcastTry_ConcurrentWithRegistration(t *testing.T) {
	b := syncx.NewBroadcaster[int]()

	const workers = 8
	var wg sync.WaitGroup

	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := "listener-" + strconv.Itoa(i)
			for range 200 {
				ch := b.RegisterBufferedListener(key, 1)
				select {
				case <-ch:
				default:
				}
				b.UnregisterListener(key)
			}
		}()
	}

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 200 {
				b.BroadcastTry(1, nil)
			}
		}()
	}

	wg.Wait()
}
