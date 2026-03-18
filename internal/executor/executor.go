package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"mb/internal/cache"
	"mb/internal/safepath"
)

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

// Run executes the plugin script. allowedRoot must be the plugin directory (or PluginsDir root);
// ExecPath is validated to be under allowedRoot before execution.
func (*Executor) Run(
	ctx context.Context,
	plugin cache.Plugin,
	args []string,
	mergedEnv []string,
	allowedRoot string,
) error {
	if plugin.ExecPath == "" {
		return errors.New("plugin has no executable path")
	}
	if err := safepath.ValidateUnderDir(plugin.ExecPath, allowedRoot); err != nil {
		return fmt.Errorf("plugin path not under allowed directory: %w", err)
	}
	command := plugin.ExecPath
	commandArgs := args
	if plugin.PluginType == "sh" {
		command = "bash"
		commandArgs = append([]string{plugin.ExecPath}, args...)
	}

	cmd := exec.CommandContext(ctx, command, commandArgs...)
	cmd.Env = mergedEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
