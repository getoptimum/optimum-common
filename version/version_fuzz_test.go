package version

import (
	"strings"
	"testing"
)

func FuzzDeriveVersion(f *testing.F) {
	seeds := []string{
		"v1.2.3+meta",
		"v1.2.3-rc1",
		"v0.0.0-20200101123456-abcdef",
		"randominput",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		out := DeriveVersion(in)
		if strings.Contains(out, "+") {
			t.Fatalf("derived version contains '+': %q", out)
		}
	})
}

func FuzzStripBuildMeta(f *testing.F) {
	f.Add("v1.2.3+meta")
	f.Fuzz(func(t *testing.T, in string) {
		out := stripBuildMeta(in)
		if strings.Contains(out, "+") {
			t.Fatalf("stripBuildMeta result contains '+': %q", out)
		}
	})
}
