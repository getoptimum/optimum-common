package jwks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/getoptimum/optimum-common/pkg/io"
	"github.com/getoptimum/optimum-common/pkg/logger"
	"github.com/getoptimum/optimum-common/pkg/net"
	"github.com/golang-jwt/jwt/v5"
)

// Config holds Cache construction params.
type Config struct {
	JWKSURL  string        //  URL of the JWKS endpoint (e.g. https://auth.optimum.com/auth/v1/.well-known/jwks.json)
	DiskPath string        // local filesystem path for the disk cache (e.g. /var/cache/jwks.json)
	Refresh  time.Duration // interval for background refreshes (e.g. 15m)
}

// Cache owns the live keyfunc and atomically swaps it on each successful refresh.
type Cache struct {
	log     logger.AppLogger // nolint:staticcheck
	url     string
	path    string
	ttl     time.Duration
	current atomic.Pointer[keyfunc.Keyfunc]
}

// New loads the cache (wire-first, disk-fallback) and starts the refresh loop.
func New(ctx context.Context, log logger.AppLogger, cfg Config) (*Cache, error) { // nolint:staticcheck
	if log == nil {
		return nil, errors.New("jwks: log is required")
	}
	if cfg.JWKSURL == "" {
		return nil, errors.New("jwks: Config.JWKSURL is required")
	}
	if cfg.DiskPath == "" {
		return nil, errors.New("jwks: Config.DiskPath is required")
	}
	if cfg.Refresh <= 0 {
		return nil, errors.New("jwks: Config.Refresh must be > 0")
	}
	c := &Cache{
		log:  log.With(logger.WithService("jwks")),
		url:  cfg.JWKSURL,
		path: cfg.DiskPath,
		ttl:  cfg.Refresh,
	}
	if err := c.load(ctx); err != nil {
		return nil, err
	}
	go c.refreshLoop(ctx)
	return c, nil
}

// Keyfunc adapts the cache to the jwt.Keyfunc signature so callers can pass
// it straight into jwt.ParseWithClaims. Safe to call concurrently with the
// refresh loop; uses an atomic load.
func (c *Cache) Keyfunc(tok *jwt.Token) (any, error) {
	kf := c.current.Load()
	if kf == nil {
		return nil, errors.New("jwks: no keyfunc loaded")
	}
	return (*kf).Keyfunc(tok)
}

// load is the initial wire-or-disk fetch at construction time. Tries the
// wire first; falls back to disk so the service can boot during a brief
// auth-provider outage.
func (c *Cache) load(ctx context.Context) error {
	if raw, err := c.fetch(ctx); err == nil {
		if writeErr := io.AtomicallySaveToFile(c.path, raw); writeErr != nil {
			// disk write failure is non-fatal — we still have the live JWKS in memory.
			c.log.Error("failed to persist JWKS to disk cache", writeErr, logger.WithString("path", c.path))
		}
		return c.swap(raw, "wire")
	} else {
		c.log.Info("JWKS wire fetch failed at boot, falling back to disk cache",
			logger.WithString("url", c.url),
			logger.WithString("path", c.path),
			logger.WithString("error", err.Error()),
		)
	}

	raw, err := io.LoadFromFile(c.path)
	if err != nil {
		return fmt.Errorf("jwks: wire fetch failed AND disk cache unavailable at %s: %w", c.path, err)
	}
	return c.swap(raw, "disk")
}

// refreshLoop is the long-interval background refresh.
func (c *Cache) refreshLoop(ctx context.Context) {
	t := time.NewTicker(c.ttl)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			raw, err := c.fetch(ctx)
			if err != nil {
				c.log.Error("JWKS refresh failed; keeping current keyfunc", err, logger.WithString("url", c.url))
				continue
			}
			if writeErr := io.AtomicallySaveToFile(c.path, raw); writeErr != nil {
				c.log.Error("failed to persist refreshed JWKS to disk cache", writeErr, logger.WithString("path", c.path))
			}
			if err := c.swap(raw, "wire-refresh"); err != nil {
				c.log.Error("refreshed JWKS could not be parsed; keeping current keyfunc", err)
			}
		}
	}
}

// fetch is one HTTP GET against the JWKS endpoint.
func (c *Cache) fetch(ctx context.Context) ([]byte, error) {
	res, code, err := net.GetCurl[json.RawMessage](ctx, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("jwks fetch: code %d: %w", code, err)
	}
	if code/100 != 2 {
		return nil, fmt.Errorf("jwks endpoint returned status %d", code)
	}
	if res == nil || len(*res) == 0 {
		return nil, errors.New("empty JWKS response body")
	}
	return *res, nil
}

// swap parses raw JWKS JSON, builds a keyfunc, and atomically replaces the
// active one. Source is "wire" / "disk" / "wire-refresh" for log lines.
func (c *Cache) swap(raw []byte, source string) error {
	kf, err := keyfunc.NewJWKSetJSON(raw)
	if err != nil {
		return fmt.Errorf("parse JWKS (%s): %w", source, err)
	}
	c.current.Store(&kf)
	c.log.Info("JWKS loaded", logger.WithString("source", source), logger.WithString("path", c.path))
	return nil
}
