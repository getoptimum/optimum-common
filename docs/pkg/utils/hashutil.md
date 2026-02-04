# Package: utils

**File:** `hashutil.go`

## Functions

### HashBytes

```go
func HashBytes(b []byte) string
```

HashBytes returns a SHA256 hash of the given bytes as a hex string.
Returns empty string if input is empty.

---

### HashSHA256

```go
func HashSHA256(data []byte) string
```

HashSHA256 computes SHA-256 hash returns hex-encoded string

---

### HashSHA256String

```go
func HashSHA256String(data []byte) [32]byte
```

HashSHA256String computes the SHA-256 hash, returns the raw 32-byte array

---

### HashSHA512

```go
func HashSHA512(data []byte) string
```

HashSHA512 computes SHA-512 hash, returns hex-encoded string

---

### HashXXHash

```go
func HashXXHash(data []byte) uint64
```

HashXXHash computes the XXHash, fast, non-cryptographic

---

### MsgHashWithTimestamp

```go
func MsgHashWithTimestamp(topic string, message []byte, timestamp int64) string
```

MsgHashWithTimestamp returns a hash of topic + message + timestamp for time-scoped identity
(e.g. duplicate detection when the same content can be sent at different times).
For content-only identity, use utils.MsgHash.

---

### WriteBool

```go
func WriteBool(h hash.Hash, v bool)
```

WriteBool writes a boolean value to the hash in a deterministic format (1 for true, 0 for false).

---

### WriteFloat32

```go
func WriteFloat32(h hash.Hash, v float32)
```

WriteFloat32 writes a float32 value to the hash by converting it to its IEEE 754 binary representation.

---

### WriteInt64

```go
func WriteInt64(h hash.Hash, v int64)
```

WriteInt64 writes an int64 value to the hash in little-endian format.
