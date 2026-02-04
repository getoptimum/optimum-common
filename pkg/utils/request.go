package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CurlConf holds configuration options for HTTP requests.
type CurlConf[T any] struct {
	Decoder func(io.Reader) error
	Client  *http.Client
}

// CurlOpts is a functional option type for configuring HTTP requests.
type CurlOpts[T any] func(*CurlConf[T])

// WithDecoder sets a custom decoder function for processing the response body.
// If set, the decoder is called instead of JSON unmarshaling.
func WithDecoder[T any](decoder func(io.Reader) error) func(*CurlConf[T]) {
	return func(c *CurlConf[T]) {
		c.Decoder = decoder
	}
}

// WithHTTPClient allows using a custom HTTP client (e.g., with keep-alive, custom timeouts)
func WithHTTPClient[T any](client *http.Client) func(*CurlConf[T]) {
	return func(c *CurlConf[T]) {
		c.Client = client
	}
}
// PatchCurl sends a PATCH request with JSON payload and returns the unmarshaled response.
// The payload is automatically marshaled to JSON.
func PatchCurl[T any](ctx context.Context, targetURL string, payload any, headers map[string]string) (res *T, statusCode int, err error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return res, 0, fmt.Errorf("unable to marshal payload: %w", err)
	}
	return CurlWithBody[T](ctx, http.MethodPatch, targetURL, payloadJSON, headers)
}

// PostCurl sends a POST request with JSON payload and returns the unmarshaled response.
// The payload is automatically marshaled to JSON.
func PostCurl[T any](ctx context.Context, targetURL string, payload any, headers map[string]string) (res *T, statusCode int, err error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return res, 0, fmt.Errorf("unable to marshal payload: %w", err)
	}
	return CurlWithBody[T](ctx, http.MethodPost, targetURL, payloadJSON, headers)
}

// CurlWithBody sends an HTTP request with the specified method, URL, JSON body, and headers.
// Returns the unmarshaled response of type T, the HTTP status code, and any error.
// Supports custom decoders and HTTP clients via options.
func CurlWithBody[T any](ctx context.Context, method, targetURL string, payloadJSON []byte, headers map[string]string, opts ...CurlOpts[T]) (res *T, statusCode int, err error) {
	config := &CurlConf[T]{}
	for _, opt := range opts {
		opt(config)
	}
	// todo inject tracer
	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return res, 0, fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return executeRequest[T](req, config)
}

// GetCurl is a generic function to send GET request with headers and return response
// expect json response
// T is a type of response, automatically unmarshalled from json
// check status code first. in some cases response can be different, so unmarshalling will fail
// status code return as is even in unmarshalling error
func GetCurl[T any](ctx context.Context, targetURL string, headers map[string]string, opts ...CurlOpts[T]) (res *T, statusCode int, err error) {
	config := &CurlConf[T]{}
	for _, opt := range opts {
		opt(config)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return res, 0, fmt.Errorf("unable to create request: %w", err)
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return executeRequest[T](req, config)
}

func executeRequest[T any](req *http.Request, config *CurlConf[T]) (res *T, statusCode int, err error) {
	client := http.DefaultClient
	if config != nil && config.Client != nil {
		client = config.Client
	}
	resp, err := client.Do(req)
	if err != nil {
		return res, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if config != nil && config.Decoder != nil {
		return nil, resp.StatusCode, config.Decoder(resp.Body)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, 0, fmt.Errorf("failed to read response body: %w", err)
	}
	var result T
	if strings.TrimSpace(strings.ReplaceAll(string(b), "\"", "")) != "" {
		if err = json.Unmarshal(b, &result); err != nil {
			return res, resp.StatusCode, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return &result, resp.StatusCode, nil
}
