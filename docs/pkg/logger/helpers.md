# Package: logger

**File:** `helpers.go`

## Functions

### WithAny

```go
func WithAny(key string, val any) Field
```

WithAny creates a Field with any value type.

---

### WithBool

```go
func WithBool(key string, v bool) Field
```

WithBool creates a Field with a boolean value.

---

### WithError

```go
func WithError(err error) *Field
```

WithError creates a Field with key "err" containing the error message.
Returns a pointer to the Field.

---

### WithFilePath

```go
func WithFilePath(filePath string) Field
```

WithFilePath creates a Field with key "file" and the given file path.

---

### WithFloat32

```go
func WithFloat32(key string, v float32) Field
```

WithFloat32 creates a Field with a float32 value.

---

### WithFloat64

```go
func WithFloat64(key string, v float64) Field
```

WithFloat64 creates a Field with a float64 value.

---

### WithFlow

```go
func WithFlow(val string) Field
```

WithFlow creates a Field with key "flow" and the given value.

---

### WithInt

```go
func WithInt(key string, v int) Field
```

WithInt creates a Field with an int value (platform-dependent size).

---

### WithInt32

```go
func WithInt32(key string, v int32) Field
```

WithInt32 creates a Field with an int32 value.

---

### WithInt64

```go
func WithInt64(key string, v int64) Field
```

WithInt64 creates a Field with an int64 value.

---

### WithModule

```go
func WithModule(val string) *Field
```

WithModule creates a Field with key "module" and the given value.
Returns a pointer to the Field.

---

### WithPeer

```go
func WithPeer(info peer.AddrInfo) Field
```

WithPeer adds the peer ID of the given AddrInfo

---

### WithPeerAddrs

```go
func WithPeerAddrs(info peer.AddrInfo) Field
```

WithPeerAddrs adds comma separated multiaddrs of the given peer

---

### WithRunID

```go
func WithRunID(val string) Field
```

WithRunID creates a Field with key "run_id" and the given value.

---

### WithService

```go
func WithService(serviceName string) Field
```

WithService creates a Field with key "service" and the given service name.

---

### WithString

```go
func WithString(key , v string) Field
```

WithString creates a Field with a string value.

---

### WithTopic

```go
func WithTopic(topic string) Field
```

WithTopic adds a pubsub topic as a string field

---

### WithTopicBytes

```go
func WithTopicBytes(topic []byte) Field
```

WithTopicBytes adds a topic represented as byte slice

---

### WithUint

```go
func WithUint(key string, v uint) Field
```

WithUint creates a Field with a uint value.
Uses Uint64 internally as slog has no Uint() function.

---

### WithUint64

```go
func WithUint64(key string, v uint64) Field
```

WithUint64 creates a Field with a uint64 value.
