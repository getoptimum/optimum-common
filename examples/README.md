# Examples

This module ships with small runnable programs that demonstrate the
packages in `optimum-common`. 
Each program lives in its own directory so multiple `main` packages do not conflict during builds.

## Running

```bash
# Authentication examples
go run ./examples/auth/parse
go run ./examples/auth/verify/mint
go run ./examples/auth/verify/server
# Local verifier: start the server then send it a minted token
go run ./examples/auth/verify/local &
TOKEN=$(go run ./examples/auth/verify/mint)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/protected

# Rate‑limit examples
go run ./examples/ratelimit/basic
go run ./examples/ratelimit/full
go run ./examples/ratelimit/walkthrough
go run ./examples/ratelimit/concurrent
```