package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getoptimum/optimum-common/auth"
)

// This example starts a tiny HTTP server that verifies bearer tokens.
func main() {
	v, err := auth.NewVerifierFromDomain("example.auth0.com", "https://api.example", nil)
	if err != nil {
		panic(err)
	}

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
