package version

import "strings"

// Version is set at build time via ldflags (e.g. -X mb/internal/version.Version=v1.0.0).
// If not set, the CLI falls back to runtime/debug.ReadBuildInfo or "dev".
var Version string

// IsReleaseBuild reports whether this binary was built with an embedded release version (GoReleaser ldflags).
// Builds from go install / make build without -X have empty Version; mb update --only-cli only applies to release builds.
func IsReleaseBuild() bool {
	return strings.TrimSpace(Version) != ""
}

// UpdateRepo is optional "owner/repo" for mb update --only-cli (GitHub). Empty uses the default upstream.
var UpdateRepo string
