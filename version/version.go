package version

import (
	"regexp"
	"runtime/debug"
	"strings"
)

type Info struct {
	Version     string // e.g., v1.2.3, v1.2.3-rc1, or "dev"
	CommitShort string // 7-char short hash or "unknown"
}

// Get reads the runtime build info
func Get() Info {
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return Info{Version: "dev", CommitShort: "unknown"}
	}
	v := deriveVersion(stripBuildMeta(bi.Main.Version))
	if v == "" {
		v = "dev"
	}
	rev := vcsRevision(bi)
	short := short(rev)
	if short == "" {
		short = "unknown"
	}
	return Info{Version: v, CommitShort: short}
}

// -----------------------------
// Version parsing (RE2-safe)
// -----------------------------

// RE2‑compatible patterns (no (?:...))
// precompiled patterns (RE2-compatible: no non-capturing groups)
var (
	// vX.Y.Z or vX.Y.Z-<prerelease>
	reExact = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)

	// vX.Y.Z-<pre>.0.<14d>-<hash>
	rePseudoWithPre = regexp.MustCompile(`^(v\d+\.\d+\.\d+-[0-9A-Za-z.-]+)\.0\.\d{14}-[0-9a-fA-F]+$`)

	// vX.Y.Z-0.<14d>-<hash>
	rePseudoBase = regexp.MustCompile(`^(v\d+\.\d+\.\d+)-0\.\d{14}-[0-9a-fA-F]+$`)

	// vX.Y.Z-YYYYMMDD-<hash>  (kept as-is; not a Go pseudo)
	reUnknown8 = regexp.MustCompile(`^v\d+\.\d+\.\d+-\d{8}-[0-9a-fA-F]+$`)

	// v0.0.0-<14d>-<hash> -> dev
	rePurePseudo = regexp.MustCompile(`^v0\.0\.0-\d{14}-[0-9a-fA-F]+$`)
)

func deriveVersion(in string) string {
	v := stripBuildMeta(in)

	// explicit pseudo-forms first
	if m := rePseudoWithPre.FindStringSubmatch(v); m != nil {
		return m[1] // e.g., v1.2.3-rc4
	}
	if m := rePseudoBase.FindStringSubmatch(v); m != nil {
		return m[1] // e.g., v1.2.3
	}

	// Pure pseudo without a base tag
	if rePurePseudo.MatchString(v) {
		return "dev"
	}

	// Unknown 8-digit date tail is preserved
	if reUnknown8.MatchString(v) {
		return v
	}

	// Exact semver (incl. prerelease) that isn’t a pseudo
	if reExact.MatchString(v) {
		return v
	}

	// catch-all
	return "dev"
}

// -----------------------------
// Helpers
// -----------------------------
func stripBuildMeta(v string) string {
	if i := strings.IndexByte(v, '+'); i >= 0 {
		return v[:i]
	}
	return v
}

func vcsRevision(bi *debug.BuildInfo) string {
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			return s.Value
		}
	}
	return ""
}

func short(h string) string {
	if len(h) >= 7 {
		return h[:7]
	}
	return h
}
