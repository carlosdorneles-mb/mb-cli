package ports

import "mb/internal/domain/plugin"

// PluginSyncStore is the persistence surface required by plugin cache sync.
type PluginSyncStore interface {
	ListPlugins() ([]plugin.Plugin, error)
	ListPluginSources() ([]plugin.PluginSource, error)
	DeleteAllPlugins() error
	DeleteAllPluginHelpGroups() error
	UpsertPluginHelpGroup(g plugin.PluginHelpGroup) error
	UpsertPlugin(p plugin.Plugin) error
	DeleteAllCategories() error
	UpsertCategory(c plugin.Category) error
}

// PluginScanner scans installed plugin trees into domain rows and help-group batches.
type PluginScanner interface {
	SetDebugLog(fn func(string))
	Scan() ([]plugin.Plugin, []plugin.Category, []plugin.ValidationWarning, [][]plugin.HelpGroupDef, error)
	ScanDir(
		localPath, installDir string,
	) ([]plugin.Plugin, []plugin.Category, []plugin.ValidationWarning, [][]plugin.HelpGroupDef, error)
}

// ShellHelperInstaller materializes embedded shell helpers under the config directory.
type ShellHelperInstaller interface {
	EnsureShellHelpers(configDir string) (string, error)
}
