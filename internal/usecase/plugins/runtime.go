package plugins

// PluginRuntime is configuration shared by plugin use cases (sync, add, remove, update).
type PluginRuntime struct {
	ConfigDir  string
	PluginsDir string
	Quiet      bool
	Verbose    bool
}
