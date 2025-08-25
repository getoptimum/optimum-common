package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"exact", "v1.2.3", "v1.2.3"},
		{"pre", "v1.2.3-rc1", "v1.2.3-rc1"},
		{"pseudo+pre", "v0.0.1-rc4.0.20250821142859-deadbeef", "v0.0.1-rc4"},
		{"pseudo base", "v1.2.3-0.20240102112233-deadbeef", "v1.2.3"},
		{"pure pseudo v000", "v0.0.0-20240102112233-deadbeef", "dev"},
		{"unknown8 keep", "v1.2.3-20250821-deadbeef", "v1.2.3-20250821-deadbeef"},
		{"devel", "(devel)", "dev"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, deriveVersion(tc.in))
		})
	}
}

// Extra focused tests

func TestDeriveVersion_PseudoPrerelease_Dirty(t *testing.T) {
	in := "v0.0.1-rc4.0.20250821142859-6128ae7a7356+dirty"
	want := "v0.0.1-rc4"
	if got := deriveVersion(in); got != want {
		t.Fatalf("deriveVersion(%q) = %q, want %q", in, got, want)
	}
}

func TestDeriveVersion_PseudoBase(t *testing.T) {
	in := "v1.2.4-0.20240102112233-abcdef1"
	want := "v1.2.4"
	if got := deriveVersion(in); got != want {
		t.Fatalf("deriveVersion(%q) = %q, want %q", in, got, want)
	}
}

func TestDeriveVersion_PurePseudo_v000(t *testing.T) {
	in := "v0.0.0-20250821142859-deadbeef"
	if got := deriveVersion(in); got != "dev" {
		t.Fatalf("deriveVersion(%q) = %q, want dev", in, got)
	}
}

func TestDeriveVersion_StripsBuildMetaFirst(t *testing.T) {
	if got := deriveVersion("v1.2.5+dirty"); got != "v1.2.5" {
		t.Fatalf("deriveVersion(strip dirty) got %q", got)
	}
	if got := deriveVersion("v1.2.5+build.meta"); got != "v1.2.5" {
		t.Fatalf("deriveVersion(strip meta) got %q", got)
	}
}
