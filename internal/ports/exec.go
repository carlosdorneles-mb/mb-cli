package ports

import (
	"context"

	"mb/internal/domain/plugin"
)

// ScriptExecutor runs a plugin entrypoint with merged environment (shell or binary).
type ScriptExecutor interface {
	Run(
		ctx context.Context,
		p plugin.Plugin,
		args []string,
		mergedEnv []string,
		allowedRoot string,
	) error
}
