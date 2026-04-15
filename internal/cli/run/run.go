package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"mb/internal/cli/runtimeflags"
	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/env"
)

// NewRunCmd returns `mb run`, which executes a subprocess with the same merged environment as plugins.
func NewRunCmd(d deps.Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command]",
		Short: "Executa um comando com as variáveis de ambiente do CLI",
		Long: `Repassa o comando e os argumentos ao executável (PATH ou caminho), com o mesmo ambiente
mesclado dos plugins (env.defaults, --env-vault, mbcli.yaml envs, ./.env no diretório atual, --env-file, --env, etc.).

As flags globais do MB (-e/--env, --env-file, --env-vault, -v/--verbose, -q/--quiet) podem ir antes
de mb (ex.: mb --env-vault st run cmd) ou logo após run (ex.: mb run --env-vault st cmd), sempre
antes do nome do executável. Flags do programa filho ficam depois do nome (ex.: mb run grep -r).

Para ajuda deste comando use: mb help run`,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				return cmd.Help()
			}
			rest, err := runtimeflags.ParseLeadingRuntimeFlags(d.Runtime, args)
			if err != nil {
				return err
			}
			if len(rest) < 1 {
				return fmt.Errorf(
					"indique o comando a executar após as flags do MB (ex.: mb run echo oi)",
				)
			}
			if _, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues); err != nil {
				return err
			}
			expanded, aliasVault, _, err := alib.ResolveForRun(
				d.Runtime.ConfigDir,
				rest[0],
				rest[1:],
			)
			if err != nil {
				return err
			}
			dRun := d
			if aliasVault != "" && d.Runtime.EnvVault == "" {
				rtCopy := *d.Runtime
				rtCopy.EnvVault = aliasVault
				dRun.Runtime = &rtCopy
			}
			merged, err := deps.BuildMergedOSEnviron(dRun, nil)
			if err != nil {
				return err
			}
			if len(expanded) < 1 {
				return fmt.Errorf("comando vazio após resolver alias")
			}
			name := expanded[0]
			bin, err := exec.LookPath(name)
			if err != nil {
				return fmt.Errorf("comando não encontrado %q: %w", name, err)
			}
			ctx := cmd.Context()
			if d.Runtime.PluginTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
				defer cancel()
			}
			c := exec.CommandContext(ctx, bin, expanded[1:]...)
			c.Env = merged
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			runErr := c.Run()
			if runErr == nil {
				return nil
			}
			var exitErr *exec.ExitError
			if errors.As(runErr, &exitErr) {
				code := exitErr.ExitCode()
				if code >= 0 {
					os.Exit(code)
				}
				os.Exit(1)
			}
			return runErr
		},
	}
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if root := c.Root(); root != nil {
			root.HelpFunc()(c, args)
		}
	})
	return cmd
}
