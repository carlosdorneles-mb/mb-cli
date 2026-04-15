package aliases

import "path/filepath"

// FilePath returns the path to aliases.yaml under the MB config directory.
func FilePath(configDir string) string {
	return filepath.Join(configDir, "aliases.yaml")
}

// ShellDir returns ~/.config/mb/shell (generated scripts).
func ShellDir(configDir string) string {
	return filepath.Join(configDir, "shell")
}
