package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	// jwt5 depends on keyfuncv2 and prev versions have vulnerabilities
	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Verifier validates tokens with JWKS + (optional) iss/aud checks.
type Verifier struct {
	jwks     *keyfunc.JWKS
	Issuer   string // e.g. "https://<domain>/"
	Audience string // expected audience (single)
}

// VerifierOptions tune JWKS refresh behavior.
type VerifierOptions struct {
	RefreshInterval     time.Duration
	RefreshRateLimit    time.Duration
	RefreshTimeout      time.Duration
	RefreshUnknownKID   bool
	HTTPClient          *http.Client
	RefreshErrorHandler func(error)
}

// NewVerifierFromDomain fetches the JWKS from
// the given domain and constructs a Verifier.
// JWKS refresh errors are surfaced via RefreshErrorHandler if set
// or logged by default.
func NewVerifierFromDomain(domain, audience string, opt *VerifierOptions) (*Verifier, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", domain)
	opts := keyfunc.Options{}
	// apply sensible defaults if nil/zero
	if opt != nil {
		if opt.RefreshInterval > 0 {
			opts.RefreshInterval = opt.RefreshInterval
		}
		if opt.RefreshRateLimit > 0 {
			opts.RefreshRateLimit = opt.RefreshRateLimit
		}
		if opt.RefreshTimeout > 0 {
			opts.RefreshTimeout = opt.RefreshTimeout
		}
		opts.RefreshUnknownKID = opt.RefreshUnknownKID

		if opt.HTTPClient != nil {
			opts.Client = opt.HTTPClient
		}
		if opt.RefreshErrorHandler != nil {
			opts.RefreshErrorHandler = opt.RefreshErrorHandler
		}
	}
	if opts.RefreshErrorHandler == nil {
		opts.RefreshErrorHandler = func(err error) { log.Printf("jwks refresh error: %v", err) }
	}

	jwks, err := keyfunc.Get(jwksURL, opts)
	if err != nil {
		return nil, fmt.Errorf("get JWKS: %w", err)
	}
	return &Verifier{
		jwks:     jwks,
		Issuer:   "https://" + domain + "/",
		Audience: audience,
	}, nil
}

// Verify parses & validates the token (signature, exp/nbf), then maps claims.
// It enforces Issuer/Audience if set on the Verifier.
func (v *Verifier) Verify(tokenStr string) (*Claims, error) {
	// Explicitly request MapClaims and validation
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
		jwt.WithAudience(v.Audience),
		jwt.WithIssuer(v.Issuer),
		jwt.WithLeeway(30*time.Second),
	)

	tok, err := parser.ParseWithClaims(tokenStr, jwt.MapClaims{}, v.jwks.Keyfunc)
	if err != nil || !tok.Valid {
		switch {
		case errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, fmt.Errorf("%w: %w", ErrInvalidIssuer, err)
		case errors.Is(err, jwt.ErrTokenInvalidAudience):
			return nil, fmt.Errorf("%w: %w", ErrInvalidAudience, err)
		default:
			return nil, fmt.Errorf("%w: %w", ErrParsingToken, err)
		}
	}

	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return FromMap(mc)
}
