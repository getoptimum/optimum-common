package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ParseUnverified extracts claims without validating the signature.
// Accepts a valid JWT and returns Claims without verifying the signature.
// Handles signature errors gracefully.
// Returns an error for completely malformed tokens.
func ParseUnverified(tokenString string) (*Claims, error) {
	tok, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		// No key: we expect signature-related error.
		return nil, nil
	})
	var mc jwt.MapClaims
	if err != nil {
		// Accept signature errors, still read claims.
		if errors.Is(err, jwt.ErrSignatureInvalid) || errors.Is(err, jwt.ErrTokenUnverifiable) {
			if tok != nil {
				if m, ok := tok.Claims.(jwt.MapClaims); ok {
					mc = m
				}
			}
		} else {
			return nil, fmt.Errorf("%w: %v", ErrParsingToken, err)
		}
	} else {
		var ok bool
		mc, ok = tok.Claims.(jwt.MapClaims)
		if !ok {
			return nil, ErrInvalidClaims
		}
	}
	if mc == nil {
		return nil, ErrInvalidClaims
	}

	return fromMap(mc)
}
