package shellhelpers

import "mb/internal/ports"

// Installer satisfies ports.ShellHelperInstaller by delegating to EnsureShellHelpers.
type Installer struct{}

// EnsureShellHelpers implements ports.ShellHelperInstaller.
func (Installer) EnsureShellHelpers(configDir string) (string, error) {
	return EnsureShellHelpers(configDir)
}

var _ ports.ShellHelperInstaller = Installer{}
