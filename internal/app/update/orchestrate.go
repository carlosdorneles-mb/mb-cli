package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	JSON                                                   bool
	// SelfUpdate, if non-nil, is used instead of SelfUpdateConfigFromDeps for the CLI check-only path (tests).
	SelfUpdate       *selfupdate.Config
	RunAllGitPlugins func(ctx context.Context) error
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
	if o.JSON && (!o.CheckOnly || !o.OnlyCLI) {
		return errors.New("use --json apenas com --only-cli --check-only")
	}

	runPlugins, runCLI, runSystem, runTools := ResolveUpdatePhases(
		o.OnlyPlugins, o.OnlyCLI, o.OnlySystem, o.OnlyTools,
	)
	toolsOnlyExclusive := o.OnlyTools && !o.OnlyPlugins && !o.OnlyCLI && !o.OnlySystem
	systemOnlyExclusive := o.OnlySystem && !o.OnlyPlugins && !o.OnlyCLI && !o.OnlyTools

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
			if o.SelfUpdate != nil {
				suCfg = o.SelfUpdate
			}
			local := strings.TrimSpace(version.Version)
			if o.JSON {
				rep, _, code, err := selfupdate.CheckOnlyDetails(ctx, suCfg, local)
				if err != nil {
					return err
				}
				raw, err := json.Marshal(rep)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(raw)); err != nil {
					return err
				}
				if code == selfupdate.ExitCodeUpdateAvailable {
					os.Exit(selfupdate.ExitCodeUpdateAvailable)
				}
			} else {
				out, code, err := selfupdate.RunCheckOnly(ctx, suCfg, local)
				if out != "" {
					LogCheckOnlyHumanLines(ctx, log, out)
				}
				if err != nil {
					return err
				}
				if code == selfupdate.ExitCodeUpdateAvailable {
					os.Exit(selfupdate.ExitCodeUpdateAvailable)
				}
			}
		} else {
			if err := RunCLIUpdate(ctx, d, log); err != nil {
				return err
			}
		}
	}

	if runSystem {
		return RunMachineUpdatePhase(ctx, cmd, log, systemOnlyExclusive)
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
