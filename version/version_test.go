package version_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/version"
	"github.com/stretchr/testify/require"
)

func TestDeriveVersion(t *testing.T) {

	cases := map[string]string{
		// exact semver
		"v1.2.3": "v1.2.3",

		// prerelease
		"v1.2.3-rc1": "v1.2.3-rc1",

		// prerelease with +dirty
		"v0.0.1-rc4.0.20250821142859-6128ae7a7356+dirty": "v0.0.1-rc4",

		// pseudo base
		"v1.2.3-0.20240102112233-deadbeef": "v1.2.3",
		"v1.2.4-0.20240102112233-abcdef1":  "v1.2.4",

		// pure pseudo (v0.0.0 -> dev)
		"v0.0.0-20240102112233-deadbeef": "dev",
		"v0.0.0-20250821142859-deadbeef": "dev",

		// unknown8 preserved (special date)
		"v1.2.3-20250821-deadbeef": "v1.2.3-20250821-deadbeef",

		// devel fallback
		"(devel)": "dev",

		// build metadata stripped
		"v1.2.5+dirty":      "v1.2.5",
		"v1.2.5+build.meta": "v1.2.5",
	}

	for in, want := range cases {
		require.Equal(t, want, version.DeriveVersion(in), "input=%q", in)
	}
}
