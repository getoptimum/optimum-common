# Authentication Integration

The Optimum projects now share a single authentication implementation provided by
[`optimum-common/auth`](./optimum-common/auth).  The module offers:

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



## Example usage

Complete runnable programs can be found under
[`optimum-common/examples/auth`](./optimum-common/examples/auth).

Run them with `go run`:

```bash
go run ./examples/auth/parse                # inspect a token
go run ./examples/auth/verify/mint          # mint a dev token
go run ./examples/auth/verify/server        # start an Auth0-backed verifier
go run ./examples/auth/verify/local         # start a local HS256 verifier
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