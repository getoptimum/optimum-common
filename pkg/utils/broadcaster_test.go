package utils_test

import (
	"sync"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestBroadcasterBroadcast(t *testing.T) {
	t.Run("deliver to two listeners", func(t *testing.T) {
		b := utils.NewBroadcaster[int]()
		ch1 := b.RegisterListener("l1")
		ch2 := b.RegisterListener("l2")

		const msg = 42

		// handshake to ensure receivers are ready before broadcasting
		start1 := make(chan struct{})
		start2 := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(2)

		var got1, got2 int

		go func() {
			defer wg.Done()
			<-start1 // unblocks when main signals
			got1 = <-ch1
		}()
		go func() {
			defer wg.Done()
			<-start2
			got2 = <-ch2
		}()

		// ensure both receivers are parked on their channel receives
		close(start1)
		close(start2)

		// broadcast on the same goroutine
		// returnS once both sends complete
		b.Broadcast(msg)

		wg.Wait()

		require.Equal(t, msg, got1, "listener l1 mismatch")
		require.Equal(t, msg, got2, "listener l2 mismatch")
	})

	t.Run("check unregister closes channel", func(t *testing.T) {
		br := utils.NewBroadcaster[int]()
		ch := br.RegisterListener("l")
		br.UnregisterListener("l")

		select {
		case _, ok := <-ch:
			require.False(t, ok, "expected closed channel after unregister")
		case <-time.After(250 * time.Millisecond):
			t.Fatal("read from unregistered channel timed out (not closed?)")
		}
	})

	t.Run("fanout preserves per-listener order", func(t *testing.T) {
		br := utils.NewBroadcaster[string]()
		chBr1 := br.RegisterListener("l1")
		chBr2 := br.RegisterListener("l2")

		messages := []string{"hello", "world", "test"}

		// handshake to ensure both receivers are ready before broadcasting
		start1 := make(chan struct{})
		start2 := make(chan struct{})

		got1 := make([]string, 0, len(messages))
		got2 := make([]string, 0, len(messages))

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			<-start1
			for i := 0; i < len(messages); i++ {
				got1 = append(got1, <-chBr1)
			}
		}()

		go func() {
			defer wg.Done()
			<-start2
			for i := 0; i < len(messages); i++ {
				got2 = append(got2, <-chBr2)
			}
		}()

		// unblock receivers
		close(start1)
		close(start2)

		// send the sequence
		// broadcast blocks until both receivers accept each message
		for _, m := range messages {
			br.Broadcast(m)
		}

		// wait for both reader goroutines to finish collecting exactly len(messages) items.
		wg.Wait()

		require.Equal(t, messages, got1, "listener l1 order/content mismatch")
		require.Equal(t, messages, got2, "listener l2 order/content mismatch")
	})

	t.Run("check no listeners does not panic", func(t *testing.T) {
		br := utils.NewBroadcaster[int]()
		// shall be a no-op (no deadlock, no panic)
		br.Broadcast(100)
	})
}
