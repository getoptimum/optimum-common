package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripBuildMeta(t *testing.T) {
	cases := map[string]string{
		"v1.2.3+dirty":      "v1.2.3",
		"v1.2.3+build.meta": "v1.2.3",
		"v1.2.3-rc1+meta":   "v1.2.3-rc1",
		"v1.2.3":            "v1.2.3",
		"":                  "",
		"+justplus":         "",
	}

	for in, want := range cases {
		got := stripBuildMeta(in)
		require.Equal(t, want, got, "input=%q", in)
	}
}

func TestShort(t *testing.T) {
	cases := map[string]string{
		"0123456789abcdef": "0123456", // trims to first 7
		"abcdef0":          "abcdef0", // exactly 7 stays as-is
		"abc":              "abc",     // shorter than 7 is unchanged
		"":                 "",        // empty unchanged (maps empty -> "unknown")
	}

	for in, want := range cases {
		got := short(in)
		require.Equal(t, want, got, "input=%q", in)
	}
}
