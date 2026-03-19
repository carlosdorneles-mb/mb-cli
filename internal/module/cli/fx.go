package cli

import (
	"go.uber.org/fx"

	"mb/internal/cli/root"
)

// CLIModule wires the root Cobra command.
var CLIModule = fx.Module("cli",
	fx.Provide(root.NewRootCmd),
)
