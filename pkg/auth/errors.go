package auth

import "errors"

var (
	ErrParsingToken    = errors.New("parsing token")
	ErrInvalidClaims   = errors.New("invalid token claims format")
	ErrInvalidIssuer   = errors.New("invalid issuer")
	ErrInvalidAudience = errors.New("invalid audience")
)
