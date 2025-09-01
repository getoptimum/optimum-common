package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ParseUnverified extracts claims w/o validating signature
func ParseUnverified(tokenString string) (*Claims, error) {
	var mc jwt.MapClaims

	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, &mc)
	if err != nil {
		// if completely malformed or not parseable.
		return nil, fmt.Errorf("%w: %w", ErrParsingToken, err)
	}

	if mc == nil {
		return nil, ErrInvalidClaims
	}
	return FromMap(mc)
}
