package plugins

import (
	"mb/internal/deps"
	"mb/internal/usecase/plugins"
)

// pluginRuntimeFromDeps builds the PluginRuntime from injected dependencies.
// Kept as a helper for commands that still use the legacy usecase functions.
func pluginRuntimeFromDeps(d deps.Dependencies) plugins.PluginRuntime {
	return plugins.PluginRuntime{
		ConfigDir:  d.Runtime.ConfigDir,
		PluginsDir: d.Runtime.PluginsDir,
		Quiet:      d.Runtime.Quiet,
		Verbose:    d.Runtime.Verbose,
	}
}
