# Package: utils

**File:** `request.go`

## Functions

### CurlWithBody

```go
func CurlWithBody[T any](ctx context.Context, method , targetURL string, payloadJSON []byte, headers map[string]string, opts ...CurlOpts[T]) (res *T, statusCode int, err error)
```

CurlWithBody sends an HTTP request with the specified method, URL, JSON body, and headers.
Returns the unmarshaled response of type T, the HTTP status code, and any error.
Supports custom decoders and HTTP clients via options.

---

### GetCurl

```go
func GetCurl[T any](ctx context.Context, targetURL string, headers map[string]string, opts ...CurlOpts[T]) (res *T, statusCode int, err error)
```

GetCurl is a generic function to send GET request with headers and return response
expect json response
T is a type of response, automatically unmarshalled from json
check status code first. in some cases response can be different, so unmarshalling will fail
status code return as is even in unmarshalling error

---

### PatchCurl

```go
func PatchCurl[T any](ctx context.Context, targetURL string, payload any, headers map[string]string) (res *T, statusCode int, err error)
```

PatchCurl sends a PATCH request with JSON payload and returns the unmarshaled response.
The payload is automatically marshaled to JSON.

---

### PostCurl

```go
func PostCurl[T any](ctx context.Context, targetURL string, payload any, headers map[string]string) (res *T, statusCode int, err error)
```

PostCurl sends a POST request with JSON payload and returns the unmarshaled response.
The payload is automatically marshaled to JSON.

---

### WithDecoder

```go
func WithDecoder[T any](decoder (io.Reader) error) (*CurlConf[T])
```

WithDecoder sets a custom decoder function for processing the response body.
If set, the decoder is called instead of JSON unmarshaling.

---

### WithHTTPClient

```go
func WithHTTPClient[T any](client *http.Client) (*CurlConf[T])
```

WithHTTPClient allows using a custom HTTP client (e.g., with keep-alive, custom timeouts)

---

## Types

### CurlConf

```go
type CurlConf[T any] struct{...}
```

CurlConf holds configuration options for HTTP requests.

---

### CurlOpts

```go
type CurlOpts[T any] (*CurlConf[T])
```

CurlOpts is a functional option type for configuring HTTP requests.
