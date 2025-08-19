package auth

//
//import (
//	"fmt"
//	"time"
//
//)
//
//a
//import (
//	"fmt"
//	"time"
//
//
//	"github.com/golang-jwt/jwt/v5"
//)
//
//// Verifier validates tokens with JWKS + (optional) iss/aud checks.
//type Verifier struct {
//	jwks     *keyfunc.JWKS
//	Issuer   string // e.g. "https://<domain>/"
//	Audience string // expected audience (single)
//}
//
//// VerifierOptions tune JWKS refresh behavior.
//type VerifierOptions struct {
//	RefreshInterval   time.Duration
//	RefreshRateLimit  time.Duration
//	RefreshTimeout    time.Duration
//	RefreshUnknownKID bool
//}
//
//func NewVerifierFromDomain(domain, audience string, opt *VerifierOptions) (*Verifier, error) {
//	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", domain)
//
//	opts := keyfunc.Options{
//		RefreshErrorHandler: func(_ error) { /* log upstream */ }, // err var unused
//	}
//	// apply sensible defaults if nil/zero
//	if opt != nil {
//		if opt.RefreshInterval > 0 {
//			opts.RefreshInterval = opt.RefreshInterval
//		}
//		if opt.RefreshRateLimit > 0 {
//			opts.RefreshRateLimit = opt.RefreshRateLimit
//		}
//		if opt.RefreshTimeout > 0 {
//			opts.RefreshTimeout = opt.RefreshTimeout
//		}
//		opts.RefreshUnknownKID = opt.RefreshUnknownKID
//	}
//
//	jwks, err := keyfunc.Get(jwksURL, opts)
//	if err != nil {
//		return nil, fmt.Errorf("get JWKS: %w", err)
//	}
//	return &Verifier{
//		jwks:     jwks,
//		Issuer:   "https://" + domain + "/",
//		Audience: audience,
//	}, nil
//}
//
//// Verify parses & validates the token (signature, exp/nbf), then maps claims.
//// It enforces Issuer/Audience if set on the Verifier.
//func (v *Verifier) Verify(tokenStr string) (*Claims, error) {
//	tok, err := jwt.Parse(tokenStr, v.jwks.Keyfunc)
//	if err != nil || !tok.Valid {
//		return nil, fmt.Errorf("jwt invalid: %w", ErrParsingToken)
//	}
//
//	mc, ok := tok.Claims.(jwt.MapClaims)
//	if !ok {
//		return nil, ErrInvalidClaims
//	}
//	// TODO: Fix this vulcheck error
//	// Standard time checks (exp/nbf) via MapClaims.Valid
//	//if err := mc.Valid(); err != nil {
//	//	return nil, fmt.Errorf("jwt time validity: %w", err)
//	//}
//
//	// Issuer check (optional)
//	if v.Issuer != "" {
//		if iss, _ := mc["iss"].(string); iss != v.Issuer {
//			return nil, ErrInvalidIssuer
//		}
//	}
//	// Audience check (optional)
//	if v.Audience != "" {
//		if !validateAudience(mc["aud"], v.Audience) {
//			return nil, ErrInvalidAudience
//		}
//	}
//
//	return fromMap(jwt.MapClaims(mc))
//}
//
//// validateAudience mirrors your proxy logic but framework-agnostic.
//func validateAudience(aud interface{}, expect string) bool {
//	switch a := aud.(type) {
//	case string:
//		return a == expect
//	case []interface{}:
//		for _, v := range a {
//			if s, ok := v.(string); ok && s == expect {
//				return true
//			}
//		}
//	}
//	return false
//}
