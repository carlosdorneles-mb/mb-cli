package version

// Version is set at build time via ldflags (e.g. -X mb/internal/version.Version=v1.0.0).
// If not set, the CLI falls back to runtime/debug.ReadBuildInfo or "dev".
var Version string
