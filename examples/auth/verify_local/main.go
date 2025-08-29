package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var sharedSecret = []byte("dev-secret-please-change")

// tiny helper to extract "Bearer <token>"
func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	const pfx = "Bearer "
	if strings.HasPrefix(h, pfx) {
		return strings.TrimSpace(h[len(pfx):])
	}
	// accept lowercase too
	const pfx2 = "bearer "
	if strings.HasPrefix(strings.ToLower(h), pfx2) {
		return strings.TrimSpace(h[len(pfx):])
	}
	return h
}

func verify(tokenString string) (*jwt.RegisteredClaims, error) {
	var claims jwt.RegisteredClaims

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		// enforce time validation using leeway to avoid clock skew flakes
		jwt.WithLeeway(30*time.Second),
	)

	_, err := parser.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return sharedSecret, nil
	})
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		tok := bearer(r)
		if tok == "" {
			http.Error(w, "missing bearer", http.StatusUnauthorized)
			return
		}
		claims, err := verify(tok)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if _, err := fmt.Fprintln(w, "hello", claims.Subject); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	addr := ":8080"
	fmt.Println("listening on", addr, "→ GET /protected (Bearer token required)")
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
