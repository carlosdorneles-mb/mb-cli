package version

// Version is set at build time via ldflags (e.g. -X mb/internal/version.Version=v1.0.0).
// If not set, the CLI falls back to runtime/debug.ReadBuildInfo or "dev".
var Version string

// UpdateRepo is optional "owner/repo" for mb self update (GitHub). Empty uses the default upstream.
var UpdateRepo string
