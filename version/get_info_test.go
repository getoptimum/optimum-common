package version_test

import (
	"regexp"
	"testing"

	"github.com/getoptimum/optimum-common/version"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	v := version.GetVersion()

	require.NotEmpty(t, v, "version must not be empty")
	require.NotContains(t, v, "+", "version should be stripped of build metadata")

	// Accept:
	//   - "dev"
	//   - semver: vX.Y.Z (optionally with -prerelease tag)
	//   - non-standard tail will be  preserved: vX.Y.Z-YYYYMMDD-<hash> (special date format set)
	semverLike := regexp.MustCompile(`^dev$|^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$|^v\d+\.\d+\.\d+-\d{8}-[0-9a-fA-F]+$`)
	require.True(t, semverLike.MatchString(v), "unexpected version format: %q", v)
}

func TestGetCommitHash(t *testing.T) {
	h := version.GetCommitHash()
	// Either 7 chars or "unknown" which is 7 chars len too
	require.Len(t, h, 7, "CommitShort must be 7 chars or 'unknown'")
}
