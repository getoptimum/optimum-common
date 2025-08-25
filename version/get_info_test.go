package version_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/version"
)

func TestGet_PublicAPI_Invariants(t *testing.T) {
	t.Parallel()
	info := version.Get()
	// Version should be non-empty, without build metadata (+...)
	require.NotEmpty(t, info.Version, "Version must not be empty")
	require.NotContains(t, info.Version, "+", "Version should be stripped of build metadata")

	// It should be "dev" or semver-like.
	semverLike := regexp.MustCompile(`^dev$|^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$|^v\d+\.\d+\.\d+-\d{8}-[0-9a-fA-F]+$`)
	require.True(t, semverLike.MatchString(info.Version), "unexpected Version format: %q", info.Version)

	// CommitShort should be 7 chars or "unknown"
	if info.CommitShort != "unknown" {
		require.Len(t, info.CommitShort, 7, "CommitShort must be 7 chars or 'unknown'")
	}
}
