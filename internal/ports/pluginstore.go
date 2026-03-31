package ports

import "mb/internal/domain/plugin"

// PluginSourceStore mutates plugin source registry rows (Git/local installs).
type PluginSourceStore interface {
	UpsertPluginSource(ps plugin.PluginSource) error
	GetPluginSource(installDir string) (*plugin.PluginSource, error)
	DeletePluginSource(installDir string) error
}

// PluginCacheStore is the full cache persistence used by sync and package management.
type PluginCacheStore interface {
	PluginSyncStore
	PluginSourceStore
}

// PluginCLIStore is PluginCacheStore plus read methods required by CLI (attach, list).
type PluginCLIStore interface {
	PluginCacheStore
	ListCategories() ([]plugin.Category, error)
	ListPluginHelpGroups() ([]plugin.PluginHelpGroup, error)
}
