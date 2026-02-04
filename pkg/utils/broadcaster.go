package utils

import "sync"

// Broadcaster is a simple implementation of a message broadcaster
// that allows multiple listeners to subscribe to messages of type T.
type Broadcaster[T any] struct {
	mu              sync.RWMutex
	messages        map[string]chan T
	activeListeners uint64
}

// NewBroadcaster creates a new broadcaster for messages of type T.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		messages: make(map[string]chan T),
	}
}

// RegisterListener registers a new listener for messages of type T.
// DO NOT FORGET call UnregisterListener when done to prevent resource leak and app block.
func (b *Broadcaster[T]) RegisterListener(key string) chan T {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.messages[key]; exists {
		return b.messages[key]
	}
	b.messages[key] = make(chan T)
	b.activeListeners++
	return b.messages[key]
}

// UnregisterListener removes a listener from the broadcaster.
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
