package deps

import (
	"go.uber.org/fx"

	mbdeps "mb/internal/deps"
)

// DepsModule bundles injected services for commands.
var DepsModule = fx.Module("deps",
	fx.Provide(mbdeps.NewDependencies),
)
