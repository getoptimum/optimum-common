# Package: identity

**File:** `identity.go`

## Functions

### EnsureIdentity

```go
func EnsureIdentity(dir string, opt ...Option) (crypto.PrivKey, error)
```

EnsureIdentity generates an identity key file in given directory.

---

### ExtractIdentityFromDir

```go
func ExtractIdentityFromDir(dir string) (*IdentityInfo, error)
```

ExtractIdentityFromDir loads an existing identity from the specified directory.
Returns an error if the identity file does not exist or cannot be read/parsed.

---

### GenIdentityEd25519

```go
func GenIdentityEd25519() (crypto.PrivKey, error)
```

GenIdentityEd25519 creates a new random Ed25519 private key.

---

### GenerateIdentitySecp256k1

```go
func GenerateIdentitySecp256k1() (crypto.PrivKey, error)
```

GenerateIdentitySecp256k1 creates a new random Secp256k1 private key.

---

## Types

### IdentityInfo

```go
type IdentityInfo struct{...}
```

IdentityInfo contains the serialized private key and peer ID.
The ID field is included to simplify integration with testing tools.

---

### Option

```go
type Option () (crypto.PrivKey, error)
```

Option is a function that generates a private key.
Used to customize key generation in EnsureIdentity.

---

## Constants

### KeyFilename

```go
const KeyFilename = "p2p.key"
```
