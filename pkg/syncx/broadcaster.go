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

// Broadcast delivers msg to every listener, blocking on each send.
func (b *Broadcaster[T]) Broadcast(msg T) {
	b.mu.RLock()
	chs := make([]chan T, 0, len(b.messages))
	for _, ch := range b.messages {
		chs = append(chs, ch)
	}
	b.mu.RUnlock()

	for _, ch := range chs {
		ch <- msg
	}
}

// BroadcastTry is non-blocking; on a full buffer it drops msg.
// onDrop runs after b.mu is released, once per drop.
func (b *Broadcaster[T]) BroadcastTry(msg T, onDrop func(key string)) {
	type target struct {
		key string
		ch  chan T
	}
	targets := make([]target, 0, len(b.messages))

	b.mu.RLock()
	for key, ch := range b.messages {
		targets = append(targets, target{key: key, ch: ch})
	}
	b.mu.RUnlock()

	type drop struct {
		key string
		n   uint64
	}
	var drops []drop

	for _, t := range targets {
		var n uint64
		select {
		case t.ch <- msg:
		default:
			n++
		}
		if onDrop != nil && n != 0 {
			drops = append(drops, drop{key: t.key, n: n})
		}
	}

	for _, d := range drops {
		for range d.n {
			onDrop(d.key)
		}
	}
}
