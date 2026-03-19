package runtime

import (
	"go.uber.org/fx"

	"mb/internal/deps"
)

// PathsModule resolves MB data directory paths and runtime shell (paths only until flags parse).
var PathsModule = fx.Module("paths",
	fx.Provide(deps.NewPaths),
	fx.Provide(NewRuntimeConfig),
	fx.Provide(NewAppConfig),
)
