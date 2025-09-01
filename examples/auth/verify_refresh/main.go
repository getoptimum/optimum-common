package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/getoptimum/optimum-common/pkg/auth"
)

// This example shows advanced verifier usage with a custom JWKS refresh handler.
func main() {
	errCh := make(chan error, 1)
	v, err := auth.NewVerifierFromDomain("example.auth0.com", "https://api.example", &auth.VerifierOptions{
		RefreshInterval:     time.Minute,
		RefreshErrorHandler: func(err error) { errCh <- err },
	})
	if err != nil {
		panic(err)
	}

	go func() {
		for err := range errCh {
			log.Printf("jwks refresh error: %v", err)
		}
	}()

	http.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		claims, err := v.Verify(token)
		if err != nil || !claims.IsActive {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if _, err := fmt.Fprintln(w, "hello", claims.Subject); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	})

	srv := &http.Server{Addr: ":8080", ReadHeaderTimeout: 5 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
