package version

import (
	"testing"
)

func TestStripBuildMeta(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"v1.2.3+dirty":      "v1.2.3",
		"v1.2.3+build.meta": "v1.2.3",
		"v1.2.3-rc1+meta":   "v1.2.3-rc1",
		"v1.2.3":            "v1.2.3",
		"":                  "",
		"+justplus":         "", // defensive: leading '+' gets stripped to empty
	}

	for in, want := range cases {
		got := stripBuildMeta(in)
		if got != want {
			t.Errorf("stripBuildMeta(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestShort(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"0123456789abcdef": "0123456", // trims to first 7
		"abcdef0":          "abcdef0", // exactly 7 stays as-is
		"abc":              "abc",     // shorter than 7 is unchanged
		"":                 "",        // empty unchanged (Get() maps empty -> "unknown")
	}

	for in, want := range cases {
		got := short(in)
		if got != want {
			t.Errorf("short(%q) = %q, want %q", in, got, want)
		}
	}
}
