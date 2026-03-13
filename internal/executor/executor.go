package executor

import (
	"context"
	"os"
	"os/exec"

	"mb/internal/cache"
)

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) Run(ctx context.Context, plugin cache.Plugin, args []string, mergedEnv []string) error {
	command := plugin.ExecPath
	commandArgs := args
	if plugin.PluginType == "sh" {
		command = "/bin/sh"
		commandArgs = append([]string{plugin.ExecPath}, args...)
	}

	cmd := exec.CommandContext(ctx, command, commandArgs...)
	cmd.Env = mergedEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
