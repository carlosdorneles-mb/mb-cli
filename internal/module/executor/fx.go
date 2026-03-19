package executor

import (
	"go.uber.org/fx"

	"mb/internal/infra/executor"
)

// ExecutorModule provides the plugin script runner.
var ExecutorModule = fx.Module("executor",
	fx.Provide(executor.New),
)
