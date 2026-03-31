package deps

import (
	"go.uber.org/fx"

	mbdeps "mb/internal/deps"
	infrakeyring "mb/internal/infra/keyring"
	"mb/internal/ports"
)

// DepsModule bundles injected services for commands.
var DepsModule = fx.Module("deps",
	fx.Provide(
		func() ports.SecretStore { return infrakeyring.SystemKeyring{} },
	),
	fx.Provide(mbdeps.NewDependencies),
)
