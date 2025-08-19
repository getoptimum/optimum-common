package version

import (
	"runtime/debug"
	"strings"
)

// Version of the binary. Set at build time using -ldflags.
var Version = ""

// CommitHash of the source code. Set at build time using -ldflags.
var CommitHash = ""

// GetLastCommitHash returns the last git commit hash embedded in build info.
// If not determined, an empty string is returned.
func GetLastCommitHash() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	// BuildInfo.Main.Version  looks like "v0.0.0-<commit>".
	data := strings.Split(strings.ReplaceAll(info.Main.Version, "+dirty", ""), "-")
	res := data[len(data)-1]
	if len(res) > 7 {
		return res[:7]
	}
	return res
}

// extractCommitHash parses commit hash from a module version string.
// Example inputs: "v0.0.0-abcdef123456", "v0.0.0-abcdef123456+dirty".
// helper to extract the commit hash from a string (to avoid mocking debug.ReadBuildInfo)
func extractCommitHash(version string) string {
	data := strings.Split(strings.ReplaceAll(version, "+dirty", ""), "-")
	res := data[len(data)-1]
	if len(res) > 7 {
		return res[:7]
	}
	return res
}
