package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripBuildMeta(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in, want string
	}{
		{"v1.2.3+dirty", "v1.2.3"},
		{"v1.2.3+build.meta", "v1.2.3"},
		{"v1.2.3-rc1+meta", "v1.2.3-rc1"},
		{"v1.2.3", "v1.2.3"},
		{"", ""},
		{"+justplus", ""}, // defensive: leading '+' gets stripped to empty
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, stripBuildMeta(tc.in), "input=%q", tc.in)
	}
}

func TestShort(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in, want string
	}{
		{"0123456789abcdef", "0123456"}, // trims to first 7
		{"abcdef0", "abcdef0"},          // exactly 7 stays as-is
		{"abc", "abc"},                  // shorter than 7 is returned unchanged
		{"", ""},                        // empty stays empty (Get() maps empty -> "unknown")
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, short(tc.in), "input=%q", tc.in)
	}
}
