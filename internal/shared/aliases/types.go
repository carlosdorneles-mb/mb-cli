package aliases

// File is the on-disk schema for ~/.config/mb/aliases.yaml.
type File struct {
	Version int              `yaml:"version"`
	Aliases map[string]Entry `yaml:"aliases"`
}

// Entry is one user-defined alias.
type Entry struct {
	Command  []string `yaml:"command"`
	EnvVault string   `yaml:"env_vault,omitempty"`
}
