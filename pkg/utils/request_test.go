package utils_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getoptimum/optimum-common/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type TestResponse struct {
	ID1 uuid.UUID `json:"id"`
	ID2 uuid.UUID `json:"id_2"`
	ID3 uuid.UUID `json:"id_3"`
}

type TestRequest struct {
	ID1 uuid.UUID `json:"id"`
	ID2 uuid.UUID `json:"id_2"`
	ID3 uuid.UUID `json:"id_3"`
}

func TestGetCurl(t *testing.T) {
	// given
	header := uuid.NewString()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	sampleResult := TestResponse{
		ID1: uuid.New(),
		ID2: uuid.New(),
		ID3: uuid.New(),
	}
	headers := map[string]string{
		"Custom": header,
	}
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, header, r.Header.Get("Custom"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
	})

	t.Run("should serve error in case of non existing address", func(t *testing.T) {
		// when
		res, code, err := utils.GetCurl[TestResponse](ctx, "http://127.0.0.1:1/test", headers)

		// then
		require.ErrorContains(t, err, "connect: connection refused")
		require.Zerof(t, code, "should return 0 code")
		require.Zero(t, res, "should return empty response")
	})
	t.Run("should serve correct request", func(t *testing.T) {
		// when
		res, code, err := utils.GetCurl[TestResponse](ctx, fmt.Sprintf("%s/test", srv.URL), headers)

		// then
		require.NoError(t, err, "should not return error")
		require.Equal(t, http.StatusOK, code, "should return 200 code")
		require.Equal(t, &sampleResult, res, "should return correct response")
	})
}

func TestPatchCurl(t *testing.T) {
	testPatchPostCurl(t, http.MethodPatch)
}

func TestPostCurl(t *testing.T) {
	testPatchPostCurl(t, http.MethodPost)
}

func testPatchPostCurl(t *testing.T, method string) {
	t.Helper()

	// given
	var (
		res  *TestResponse
		code int
		err  error
	)
	header := uuid.NewString()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	sampleResult := TestResponse{
		ID1: uuid.New(),
		ID2: uuid.New(),
		ID3: uuid.New(),
	}
	sampleRequest := TestRequest{
		ID1: uuid.New(),
		ID2: uuid.New(),
		ID3: uuid.New(),
	}
	headers := map[string]string{
		"Custom": header,
	}
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, method, r.Method)
		require.Equal(t, header, r.Header.Get("Custom"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req TestRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, sampleRequest, req)

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
	})
	t.Run("should serve error in case of non existing address", func(t *testing.T) {
		// when
		switch method {
		case http.MethodPatch:
			res, code, err = utils.PatchCurl[TestResponse](ctx, "http://127.0.0.1:1/test", sampleRequest, headers)
		case http.MethodPost:
			res, code, err = utils.PostCurl[TestResponse](ctx, "http://127.0.0.1:1/test", sampleRequest, headers)
		}

		// then
		require.ErrorContains(t, err, "connect: connection refused")
		require.Zerof(t, code, "should return 0 code")
		require.Zero(t, res, "should return empty response")
	})
	t.Run("should serve correct request", func(t *testing.T) {
		// when
		switch method {
		case http.MethodPatch:
			res, code, err = utils.PatchCurl[TestResponse](ctx, fmt.Sprintf("%s/test", srv.URL), sampleRequest, headers)
		case http.MethodPost:
			res, code, err = utils.PostCurl[TestResponse](ctx, fmt.Sprintf("%s/test", srv.URL), sampleRequest, headers)
		}

		// then
		require.NoError(t, err, "should not return error")
		require.Equal(t, http.StatusOK, code, "should return 200 code")
		require.Equal(t, &sampleResult, res, "should return correct response")
	})
}

func TestWithHTTPClient(t *testing.T) {
	// given
	header := uuid.NewString()
	sampleResult := TestResponse{
		ID1: uuid.New(),
		ID2: uuid.New(),
		ID3: uuid.New(),
	}
	sampleRequest := TestRequest{
		ID1: uuid.New(),
		ID2: uuid.New(),
		ID3: uuid.New(),
	}
	headers := map[string]string{
		"Custom": header,
	}

	t.Run("GetCurl with custom client timeout should fail on timeout", func(t *testing.T) {
		// given
		ctx := context.Background()
		// Create a server that delays response
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond) // Delay longer than client timeout
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
		})

		// Create custom client with very short timeout
		customClient := &http.Client{
			Timeout: 50 * time.Millisecond,
		}

		// when
		res, code, err := utils.GetCurl[TestResponse](
			ctx,
			fmt.Sprintf("%s/test", srv.URL),
			headers,
			utils.WithHTTPClient[TestResponse](customClient),
		)

		// then
		require.Error(t, err, "should return error due to timeout")
		require.True(t, containsTimeoutError(err), "error should be timeout-related: %v", err)
		require.Zerof(t, code, "should return 0 code")
		require.Zero(t, res, "should return empty response")
	})

	t.Run("GetCurl with custom client should succeed with sufficient timeout", func(t *testing.T) {
		// given
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, header, r.Header.Get("Custom"))
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
		})

		// Create custom client with sufficient timeout
		customClient := &http.Client{
			Timeout: 500 * time.Millisecond,
		}

		// when
		res, code, err := utils.GetCurl[TestResponse](
			ctx,
			fmt.Sprintf("%s/test", srv.URL),
			headers,
			utils.WithHTTPClient[TestResponse](customClient),
		)

		// then
		require.NoError(t, err, "should not return error")
		require.Equal(t, http.StatusOK, code, "should return 200 code")
		require.Equal(t, &sampleResult, res, "should return correct response")
	})

	t.Run("CurlWithBody with custom client timeout should fail on timeout", func(t *testing.T) {
		// given
		ctx := context.Background()
		// Create a server that delays response
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond) // Delay longer than client timeout
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
		})

		// Create custom client with very short timeout
		customClient := &http.Client{
			Timeout: 50 * time.Millisecond,
		}

		payloadJSON, err := json.Marshal(sampleRequest)
		require.NoError(t, err)

		// when
		res, code, err := utils.CurlWithBody[TestResponse](
			ctx,
			http.MethodPost,
			fmt.Sprintf("%s/test", srv.URL),
			payloadJSON,
			headers,
			utils.WithHTTPClient[TestResponse](customClient),
		)

		// then
		require.Error(t, err, "should return error due to timeout")
		require.True(t, containsTimeoutError(err), "error should be timeout-related: %v", err)
		require.Zerof(t, code, "should return 0 code")
		require.Zero(t, res, "should return empty response")
	})

	t.Run("CurlWithBody with custom client should succeed with sufficient timeout", func(t *testing.T) {
		// given
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, header, r.Header.Get("Custom"))
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req TestRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			require.Equal(t, sampleRequest, req)

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
		})

		// Create custom client with sufficient timeout
		customClient := &http.Client{
			Timeout: 500 * time.Millisecond,
		}

		payloadJSON, err := json.Marshal(sampleRequest)
		require.NoError(t, err)

		// when
		res, code, err := utils.CurlWithBody[TestResponse](
			ctx,
			http.MethodPost,
			fmt.Sprintf("%s/test", srv.URL),
			payloadJSON,
			headers,
			utils.WithHTTPClient[TestResponse](customClient),
		)

		// then
		require.NoError(t, err, "should not return error")
		require.Equal(t, http.StatusOK, code, "should return 200 code")
		require.Equal(t, &sampleResult, res, "should return correct response")
	})

	t.Run("GetCurl without custom client should use default client", func(t *testing.T) {
		// given
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, header, r.Header.Get("Custom"))
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(sampleResult))
		})

		// when - no custom client provided
		res, code, err := utils.GetCurl[TestResponse](
			ctx,
			fmt.Sprintf("%s/test", srv.URL),
			headers,
		)

		// then
		require.NoError(t, err, "should not return error")
		require.Equal(t, http.StatusOK, code, "should return 200 code")
		require.Equal(t, &sampleResult, res, "should return correct response")
	})
}

func containsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(strings.ToLower(errStr), "timeout") ||
		strings.Contains(strings.ToLower(errStr), "deadline exceeded")
}

func newTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}
