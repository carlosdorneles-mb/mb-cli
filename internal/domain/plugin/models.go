// Package plugin holds domain types for the plugin subsystem (cache rows, sources, help metadata).
package plugin

// Category describes a plugin category row in the cache.
type Category struct {
	Path        string
	Description string
	ReadmePath  string
	Hidden      bool
	GroupID     string // help group for nested categories (e.g. infra/k8s → INFRAESTRUTURA)
	AliasesJSON string // JSON array of strings for Cobra Aliases on the category command
}

// Plugin describes a leaf plugin command row in the cache.
type Plugin struct {
	ID              int64
	CommandPath     string // e.g. "infra/ci/deploy"
	CommandName     string
	Description     string
	ExecPath        string // empty for flags-only
	PluginType      string // "sh"|"bin" or "" for flags-only
	ConfigHash      string
	ReadmePath      string
	FlagsJSON       string // for flags-only: JSON map of flag name -> {type, entrypoint}
	UseTemplate     string // Cobra Use (optional)
	ArgsCount       int    // Cobra ExactArgs (0 = no validation)
	AliasesJSON     string // JSON array of strings for Cobra Aliases
	Example         string
	LongDescription string
	Deprecated      string
	PluginDir       string // absolute path to plugin directory (manifest folder); for execution root
	Hidden          bool
	EnvFilesJSON    string // manifest env_files as JSON array of {file, group}
	GroupID         string // help group for nested leaves; empty = default COMANDOS
}

// PluginHelpGroup is a Cobra help section for nested plugin commands (from groups.yaml).
type PluginHelpGroup struct {
	GroupID string
	Title   string
}

// PluginSource represents one installation and its git origin/version or local path.
type PluginSource struct {
	InstallDir string
	GitURL     string
	RefType    string // "tag" | "branch"
	Ref        string
	Version    string
	LocalPath  string
	UpdatedAt  string
}
