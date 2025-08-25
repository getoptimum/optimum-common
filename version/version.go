package version

import (
	"regexp"
	"runtime/debug"
	"strings"
)

// Public: stable header other repos use
const HeaderCLIVersion = "X-CLI-Version"

// Public globals (populated at init from build info).
var (
	// Version is the last semver-like tag, or "dev" if none.
	Version = "dev"
	// LastCommitHashShort is the 7-char abbreviated commit hash if available.
	LastCommitHashShort = "unknown"
)

// Allow tests to stub build info.
var readBuildInfo = debug.ReadBuildInfo

// =====================
// Build-info bootstrap
// =====================

func loadFromBuildInfo(bi *debug.BuildInfo) {
	if bi == nil {
		return
	}
	Version = deriveVersion(bi.Main.Version)
	if rev := vcsRevision(bi); rev != "" {
		LastCommitHashShort = short(rev)
	}
}

func init() {
	if bi, ok := readBuildInfo(); ok {
		loadFromBuildInfo(bi)
	}
}

// =====================
// Version parsing (RE2)
// =====================

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

// =====================
//  Helpers
// ====================

func stripBuildMeta(v string) string {
	// remove trailing "+dirty" or any "+<meta>"
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
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
	if h == "" {
		return "unknown"
	}
	return h
}
