package plugin

import "os"

// DefaultPluginSubDir is the default subdirectory name where plugins
// are expected to live inside a repository or package root.
const DefaultPluginSubDir = "src"

// PluginSubDirEnv is the environment variable that overrides the default
// subdirectory for plugin discovery.
//
// If unset or empty string, DefaultPluginSubDir ("src") is used.
// If set to an empty string explicitly (MB_PLUGIN_SUBDIR=),
// subdirectory detection is disabled and plugins are scanned from root.
const PluginSubDirEnv = "MB_PLUGIN_SUBDIR"

// SubDir returns the effective plugin subdirectory.
// It reads MB_PLUGIN_SUBDIR from the environment; if not set, returns "src".
// Returns empty string when the env var is explicitly set to empty (disables subdir detection).
func SubDir() string {
	if v, ok := os.LookupEnv(PluginSubDirEnv); ok {
		return v // may be "" — disables subdir detection
	}
	return DefaultPluginSubDir
}
