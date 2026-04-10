package infra

import (
	"go.uber.org/fx"

	mbfs "mb/internal/infra/fs"
	"mb/internal/infra/plugins"
	"mb/internal/infra/shellhelpers"
	"mb/internal/ports"
)

// InfraModule provides concrete implementations for infrastructure ports.
// These are the "real" adapters that talk to the OS, Git, and shell helpers.
var InfraModule = fx.Module("infra",
	fx.Provide(
		func() ports.Filesystem { return mbfs.OS{} },
		func() ports.GitOperations { return plugins.GitService{} },
		func() ports.ShellHelperInstaller { return shellhelpers.Installer{} },
		func() ports.PluginLayoutValidator { return plugins.LayoutValidator{} },
	),
)
