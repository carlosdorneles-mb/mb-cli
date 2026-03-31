package plugins

import (
	"go.uber.org/fx"

	"mb/internal/deps"
	infraplugins "mb/internal/infra/plugins"
	"mb/internal/ports"
)

func newScanner(p *deps.Paths) *infraplugins.Scanner {
	return infraplugins.NewScanner(p.PluginsDir)
}

// PluginsModule provides the plugin filesystem scanner.
var PluginsModule = fx.Module("plugins",
	fx.Provide(
		newScanner,
		func(s *infraplugins.Scanner) ports.PluginScanner { return s },
	),
)
