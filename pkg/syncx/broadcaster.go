package syncx

import "sync"

// Broadcaster fans out messages of type T to named listeners.
type Broadcaster[T any] struct {
	mu              sync.RWMutex
	messages        map[string]*listener[T]
	activeListeners uint64
}

type listener[T any] struct {
	ch chan T
	mu sync.Mutex // serializes BroadcastTry on this channel
}

// NewBroadcaster creates a new broadcaster for messages of type T.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		messages: make(map[string]*listener[T]),
	}
}

// RegisterListener registers an unbuffered listener.
// Call UnregisterListener when done to avoid leaks and blocking Broadcast.
func (b *Broadcaster[T]) RegisterListener(key string) chan T {
	return b.registerListener(key, 0)
}

// RegisterBufferedListener registers a listener with a bounded channel.
func (b *Broadcaster[T]) RegisterBufferedListener(key string, bufSize int) chan T {
	if bufSize <= 0 {
		bufSize = 1
	}
	return b.registerListener(key, bufSize)
}

func (b *Broadcaster[T]) registerListener(key string, bufSize int) chan T {
	b.mu.Lock()
	defer b.mu.Unlock()
	if l, exists := b.messages[key]; exists {
		return l.ch
	}
	var ch chan T
	if bufSize == 0 {
		ch = make(chan T)
	} else {
		ch = make(chan T, bufSize)
	}
	b.messages[key] = &listener[T]{ch: ch}
	b.activeListeners++
	return ch
}

// UnregisterListener removes a listener and closes its channel.
func (b *Broadcaster[T]) UnregisterListener(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if l, ok := b.messages[key]; ok {
		close(l.ch)
		delete(b.messages, key)
		b.activeListeners--
	}
}

// Broadcast delivers msg to every listener, blocking on each send.
func (b *Broadcaster[T]) Broadcast(msg T) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, l := range b.messages {
		l.ch <- msg
	}
}

// BroadcastTry delivers msg without blocking. onDrop is called once per dropped
// value for that listener (oldest evicted first, then the new value if still full).
func (b *Broadcaster[T]) BroadcastTry(msg T, onDrop func(key string)) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for key, l := range b.messages {
		l.mu.Lock()
		n := trySendDropOldest(l.ch, msg)
		l.mu.Unlock()
		if onDrop == nil {
			continue
		}
		for range n {
			onDrop(key)
		}
	}
}

func trySendDropOldest[T any](ch chan T, v T) (dropped uint64) {
	select {
	case ch <- v:
		return 0
	default:
	}
	select {
	case <-ch:
		dropped++
	default:
	}
	select {
	case ch <- v:
	default:
		dropped++
	}
	return dropped
}
