package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"mb/internal/deps"
)

// NewRunCmd returns `mb run`, which executes a subprocess with the same merged environment as plugins.
func NewRunCmd(d deps.Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command]",
		Short: "Executa um comando com as variáveis de ambiente do CLI",
		Long: `Repassa o comando e os argumentos ao executável (PATH ou caminho), com o mesmo ambiente
mesclado dos plugins (env.defaults, --env-group, ./.env no diretório atual, --env-file, --env, etc.).

Para ajuda deste comando use: mb help run`,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				return cmd.Help()
			}
			merged, err := deps.BuildMergedOSEnviron(d, nil)
			if err != nil {
				return err
			}
			name := args[0]
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
			c := exec.CommandContext(ctx, bin, args[1:]...)
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
