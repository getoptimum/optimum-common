package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := []byte("dev-secret-please-change")
	now := time.Now()

	claims := jwt.MapClaims{
		"iss": "http://localhost/dev",
		"aud": "https://api.example",
		"sub": "Valentyn",
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Minute).Unix(),
		// add any custom claims your verifier maps to IsActive (if needed)
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(secret)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
