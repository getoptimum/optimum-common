# Package: entities

**File:** `p2p_messages.go`

## Functions

### UnmarshalP2PMessage

```go
func UnmarshalP2PMessage(data []byte) (*P2PMessage, error)
```

UnmarshalP2PMessage deserializes JSON data into a P2PMessage struct.

---

## Types

### P2PMessage

```go
type P2PMessage struct{...}
```

**Methods:**

- `DecodeFrom` - DecodeFrom reads JSON data from an io.Reader and decodes it into the P2PMessage struct.
- `Marshal` - Marshal serializes the P2PMessage into JSON format.
