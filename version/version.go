package version

// Version of the binary. Set at build time using -ldflags.
var Version = "unknown"

// CommitHash : short version of git commit hash of the source code. Set at build time using -ldflags.
var CommitHash = "unknown"

// HeaderCLIVersion is the HTTP header used to identify CLI versions.
const HeaderCLIVersion = "X-CLI-Version"
