package executor

import (
	"go.uber.org/fx"

	infraexec "mb/internal/infra/executor"
	"mb/internal/ports"
)

// ExecutorModule provides the plugin script runner.
var ExecutorModule = fx.Module("executor",
	fx.Provide(
		infraexec.New,
		func(e *infraexec.Executor) ports.ScriptExecutor { return e },
	),
)
