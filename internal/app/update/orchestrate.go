package update

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/infra/selfupdate"
	"mb/internal/shared/system"
	"mb/internal/shared/version"
)

// Options configures phased mb update execution.
type Options struct {
	OnlyPlugins, OnlyCLI, OnlySystem, OnlyTools, CheckOnly bool
	RunAllGitPlugins                                       func(ctx context.Context) error
}

// Run executes enabled update phases in order (plugins, tools, CLI, system).
func Run(
	ctx context.Context,
	cmd *cobra.Command,
	d deps.Dependencies,
	log *system.Logger,
	o Options,
) error {
	if o.CheckOnly && !o.OnlyCLI {
		return errors.New("use --check-only apenas com --only-cli")
	}

	runPlugins, runCLI, runSystem, runTools := ResolveUpdatePhases(
		o.OnlyPlugins, o.OnlyCLI, o.OnlySystem, o.OnlyTools,
	)
	toolsOnlyExclusive := o.OnlyTools && !o.OnlyPlugins && !o.OnlyCLI && !o.OnlySystem

	if runPlugins {
		if err := o.RunAllGitPlugins(ctx); err != nil {
			return err
		}
	}

	if runTools {
		if err := RunToolsUpdateAllPhase(ctx, cmd, log, toolsOnlyExclusive); err != nil {
			return err
		}
	}

	if runCLI {
		if o.CheckOnly {
			suCfg := SelfUpdateConfigFromDeps(d)
			local := strings.TrimSpace(version.Version)
			out, code, err := selfupdate.RunCheckOnly(ctx, suCfg, local)
			if out != "" {
				LogInfoLines(ctx, log, out)
			}
			if err != nil {
				return err
			}
			if code == selfupdate.ExitCodeUpdateAvailable {
				os.Exit(selfupdate.ExitCodeUpdateAvailable)
			}
		} else {
			if err := RunCLIUpdate(ctx, d, log); err != nil {
				return err
			}
		}
	}

	if runSystem {
		return RunSystemUpdate(ctx, log)
	}
	return nil
}

// ResolveUpdatePhases returns which phases to run. If no --only-* flag is set, all four run.
func ResolveUpdatePhases(
	onlyPlugins, onlyCLI, onlySystem, onlyTools bool,
) (plugins, cli, system, tools bool) {
	if !onlyPlugins && !onlyCLI && !onlySystem && !onlyTools {
		return true, true, true, true
	}
	return onlyPlugins, onlyCLI, onlySystem, onlyTools
}
