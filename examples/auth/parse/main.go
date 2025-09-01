package main

import (
	"fmt"

	"github.com/getoptimum/optimum-common/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
)

// This example shows how to inspect a token without verifying its signature.
func main() {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "alice"}).SignedString([]byte("secret"))
	claims, err := auth.ParseUnverified(token)
	if err != nil {
		panic(err)
	}
	fmt.Println("subject:", claims.Subject)
}
