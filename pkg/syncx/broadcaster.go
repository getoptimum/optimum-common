package syncx

import "sync"

// Broadcaster fans out messages of type T to named listeners.
type Broadcaster[T any] struct {
	mu              sync.RWMutex
	messages        map[string]chan T
	activeListeners uint64
}

// NewBroadcaster creates a new broadcaster for messages of type T.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{messages: make(map[string]chan T)}
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
	if ch, exists := b.messages[key]; exists {
		return ch
	}
	var ch chan T
	if bufSize == 0 {
		ch = make(chan T)
	} else {
		ch = make(chan T, bufSize)
	}
	b.messages[key] = ch
	b.activeListeners++
	return ch
}

// UnregisterListener removes a listener and closes its channel.
func (b *Broadcaster[T]) UnregisterListener(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, ok := b.messages[key]; ok {
		close(ch)
		delete(b.messages, key)
		b.activeListeners--
	}
}

// Broadcast sends a message to all registered listeners.
func (b *Broadcaster[T]) Broadcast(msg T) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for i := range b.messages {
		b.messages[i] <- msg
	}
}

// BroadcastTry is non-blocking; on a full buffer it drops msg.
// onDrop runs after b.mu is released, once per listener with a drop count.
func (b *Broadcaster[T]) BroadcastTry(msg T, onDrop func(key string, num uint64)) {
	drops := make(map[string]uint64, b.activeListeners)

	b.mu.Lock()
	for key, ch := range b.messages {
		select {
		case ch <- msg:
		default:
			drops[key] = 1
		}
	}
	b.mu.Unlock()

	if onDrop == nil {
		return
	}
	for key, n := range drops {
		onDrop(key, n)
	}
}
