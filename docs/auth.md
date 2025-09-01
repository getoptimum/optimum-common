# Authentication Integration

The Optimum projects now share a single authentication implementation provided by
[`optimum-common/examples/auth`](../examples/auth).

- A unified `Claims` model used by all services and clients.
- `ParseUnverified` for reading JWTs without validating the signature (used by
  CLI tools when inspecting locally stored tokens).
- `Verifier` which validates tokens against Auth0 JWKS with optional issuer and
  audience checks.

## Repository changes
### mump2p-cli
- Replace custom `TokenClaims` and token parser with wrappers around the shared
  module.
- CLI commands call `auth.ParseUnverified` to inspect tokens.
- Rate limiter operate directly on the common `Claims` type. (TODO: resolve with ratelimit package)

### optimum-proxy
- Remove  JWT parsing and JWKS handling in the middleware.
- Middleware and API handlers rely on `auth.Verifier` and `auth.Claims`.
- WebSocket helpers parse tokens via `auth.ParseUnverified` when auth is
  disabled.

### optimum-p2p
- Delete local JWKS cache and token parser; the service uses
  `auth.NewVerifierFromDomain` for validation and `auth.ParseUnverified` for
  claim extraction.
- Usage tracking works with the shared `Claims` structure.


### How to use the examples

#### mint – generate a development token

```bash
go run ./examples/auth/mint > token.jwt
cat token.jwt
```

Use the printed token with the `verify_local` example.

#### parse – inspect a token without verifying

```bash
go run ./examples/auth/parse
```

Expected output:

```
subject: alice
```

#### verify – Auth0-backed verification service

```bash
go run ./examples/auth/verify
```

In another terminal, call the protected endpoint with a valid token:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/protected
```

#### verify_local – verify HS256 tokens locally

```bash
go run ./examples/auth/verify_local &
TOKEN=$(go run ./examples/auth/mint)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/protected
```

#### verify_refresh – custom JWKS refresh handler

```bash
go run ./examples/auth/verify_refresh
```

Refresh failures are logged via the `RefreshErrorHandler`.

## Example usage

Complete runnable programs can be found under
[`optimum-common/examples/auth`](./optimum-common/examples/auth).

Run them with `go run`:

```bash
go run ./examples/auth/mint               # mint a dev token
go run ./examples/auth/parse              # inspect a token
go run ./examples/auth/verify             # start an Auth0-backed verifier
go run ./examples/auth/verify_local       # start a local HS256 verifier
go run ./examples/auth/verify_refresh     # advanced: custom JWKS refresh handler       # start a local HS256 verifier
```

### CLI – inspect a token without verifying the signature

```go
//token, _ := os.ReadFile("token.jwt")
//claims, _ := auth.ParseUnverified(string(token))
//fmt.Println("client", claims.ClientID)
token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user"}).SignedString([]byte("secret"))
claims, _ := auth.ParseUnverified(token)
fmt.Println("client", claims.Subject)
```

### HTTP service – verify a bearer token

```go
v, _ := auth.NewVerifierFromDomain("example.auth0.com", "https://api.example", nil)
http.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
    token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
    claims, err := v.Verify(token)
    if err != nil || !claims.IsActive {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    w.Write([]byte("hello " + claims.Subject))
})
```

## Migration notes

- Each repository imports `github.com/getoptimum/optimum-common/auth`.
- Existing environment variables for Auth0 domain and audience remain
  unchanged.
- Any custom JWT parsing or claim structs can be removed in favour of the
  shared package.


## Errors

The package surfaces errors wrapped with `fmt.Errorf` for `errors.Is` checks:
- `ErrParsingToken`
- `ErrInvalidClaims`
- `ErrInvalidIssuer`
- `ErrInvalidAudience`

## Security considerations
- Only RSA algorithms (`RS256`, `RS384`, `RS512`) are accepted. (should enforce)
- `exp` and `nbf` claims are validated with a 30s leeway to tolerate clock skew.
- Provide a `RefreshErrorHandler` to monitor JWKS refresh failures and rotate keys.