package version

import "testing"

func TestDeriveVersion_RegexTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		// --- exact semver (with/without prerelease) ---
		{"exact semver", "v1.2.3", "v1.2.3"},
		{"exact prerelease", "v1.2.3-rc4", "v1.2.3-rc4"},
		{"exact prerelease with build meta", "v1.2.3-rc1+build.meta", "v1.2.3-rc1"},
		{"exact semver with build meta", "v1.2.3+meta.info", "v1.2.3"},

		// --- pseudo with prerelease: vX.Y.Z-<pre>.0.<14d>-<hash>(+dirty?) -> base tag with prerelease ---
		{"pseudo prerelease +dirty (upper hex ok)", "v0.0.1-rc4.0.20250821142859-6128AE7a7356+dirty", "v0.0.1-rc4"},
		{"pseudo prerelease", "v2.3.4-beta.0.20240102112233-abcdef1", "v2.3.4-beta"},

		// --- pseudo with base tag: vX.Y.Z-0.<14d>-<hash> -> base tag ---
		{"pseudo base tag -0.", "v1.2.3-0.20240102112233-ABCDEF1", "v1.2.3"},

		// --- pure pseudo without base: v0.0.0-<14d>-<hash> -> dev ---
		{"pure pseudo v0.0.0 -> dev", "v0.0.0-20240102112233-abcdef1", "dev"},

		// --- unknown tails we intentionally keep as-is (not Go pseudo) ---
		{"unknown 8-digit tail kept", "v1.2.3-20250821-abcdef", "v1.2.3-20250821-abcdef"},

		// --- junk / non-semver-ish -> dev ---
		{"junk (devel)", "(devel)", "dev"},
		{"too short v1.2", "v1.2", "dev"},

		// --- guard: dots inside core must not be mistaken for pseudo markers ---
		{"not pseudo: core has .0. segments", "v0.0.1", "v0.0.1"},

		// --- corner: timestamp wrong length -> not matched as pseudo; falls back to exact if it fits ---
		// This one is not a Go pseudo-version (13 digits); since it still matches "exact with prerelease",
		// deriveVersion should return it unchanged.
		{"bad pseudo timestamp length -> unchanged exact", "v1.2.3-rc4.0.2025082114285-abcdef", "v1.2.3-rc4.0.2025082114285-abcdef"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := deriveVersion(tc.in)
			if got != tc.want {
				t.Fatalf("deriveVersion(%q) = %q, want %q", tc.in, got, tc.want)
			}
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
	in := "v1.2.3-0.20240102112233-abcdef1"
	want := "v1.2.3"
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
	if got := deriveVersion("v1.2.3+dirty"); got != "v1.2.3" {
		t.Fatalf("deriveVersion(strip dirty) got %q", got)
	}
	if got := deriveVersion("v1.2.3+build.meta"); got != "v1.2.3" {
		t.Fatalf("deriveVersion(strip meta) got %q", got)
	}
}
