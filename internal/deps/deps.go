package deps

import (
	"time"

	"mb/internal/ports"
	"mb/internal/shared/config"
)

// RuntimeConfig combines resolved Paths with CLI/runtime flags.
type RuntimeConfig struct {
	Paths
	Verbose     bool
	Quiet       bool
	EnvFilePath string
	// EnvVault overlays ~/.config/mb/.env.<EnvVault> on env.defaults when running plugins.
	EnvVault        string
	InlineEnvValues []string
	// PluginTimeout limits how long a plugin script can run. Zero means no limit.
	PluginTimeout time.Duration
}

// Dependencies groups services injected into commands.
type Dependencies struct {
	Runtime     *RuntimeConfig
	AppConfig   config.AppConfig
	Store       ports.PluginCLIStore
	Scanner     ports.PluginScanner
	Executor    ports.ScriptExecutor
	SecretStore ports.SecretStore
	OnePassword ports.OnePasswordEnv // optional; nil disables 1Password env integration
}

// NewDependencies constructs the dependency bundle for Fx / tests.
func NewDependencies(
	runtime *RuntimeConfig,
	appCfg config.AppConfig,
	store ports.PluginCLIStore,
	scanner ports.PluginScanner,
	exec ports.ScriptExecutor,
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
) Dependencies {
	return Dependencies{
		Runtime:     runtime,
		AppConfig:   appCfg,
		Store:       store,
		Scanner:     scanner,
		Executor:    exec,
		SecretStore: secrets,
		OnePassword: onePassword,
	}
}
