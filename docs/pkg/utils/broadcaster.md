# Package: utils

**File:** `broadcaster.go`

## Functions

### NewBroadcaster

```go
func NewBroadcaster[T any]() *Broadcaster[T]
```

NewBroadcaster creates a new broadcaster for messages of type T.

---

## Types

### Broadcaster

```go
type Broadcaster[T any] struct{...}
```

Broadcaster is a simple implementation of a message broadcaster
that allows multiple listeners to subscribe to messages of type T.

**Methods:**

- `Broadcast` - Broadcast sends a message to all registered listeners.
- `RegisterListener` - RegisterListener registers a new listener for messages of type T.
- `UnregisterListener` - UnregisterListener removes a listener from the broadcaster.
