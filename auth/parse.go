package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func ParseUnverified(tokenString string) (*Claims, error) {
	var mc jwt.MapClaims

	// Use the dedicated ParseUnverified API (doesn't validate signature at all).
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, &mc)
	if err != nil {
		// Token is completely malformed or not parseable.
		return nil, fmt.Errorf("%w: %w", ErrParsingToken, err)
	}

	// Ensure we actually got claims
	if mc == nil {
		return nil, ErrInvalidClaims
	}
	return fromMap(mc)
}
