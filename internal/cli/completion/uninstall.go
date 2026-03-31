package completion

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/shared/system"
)

// NewUninstallCmd regista `completion uninstall` no comando completion pai.
func NewUninstallCmd() *cobra.Command {
	var (
		shellFlag string
		rcFile    string
		dryRun    bool
		yes       bool
	)

	cmd := &cobra.Command{
		Use: "uninstall",
		// Short/Long em português em root.customizeCompletionPT.
		Short: "Remove shell completion block from your profile file",
		Long: `Detects SHELL or uses --shell, reads the profile file and removes the
mb-cli completion block if present. No-op if the file or block is missing.

Non-interactive sessions require --yes (or use --dry-run).`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			shell := shellFlag
			var err error
			if shell == "" {
				shell, err = DetectShell()
				if err != nil {
					return err
				}
			} else {
				shell, err = NormalizeShellName(shell)
				if err != nil {
					return err
				}
			}

			rcPath, err := ProfilePath(shell, rcFile)
			if err != nil {
				return err
			}

			existing, err := os.ReadFile(rcPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintf(cmd.OutOrStdout(), "Nada a remover: %s não existe.\n", rcPath)
					return nil
				}
				return fmt.Errorf("ler %s: %w", rcPath, err)
			}

			s := string(existing)
			if !strings.Contains(s, BlockBegin) {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"Nada a remover: nenhum bloco mb-cli em %s.\n",
					rcPath,
				)
				return nil
			}

			newContent := RemoveCompletionBlock(s, BlockBegin, BlockEnd)
			if newContent == s {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"Nada a remover: marcadores incompletos ou bloco ausente em %s.\n",
					rcPath,
				)
				return nil
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Ficheiro: %s\n\n%s", rcPath, newContent)
				return nil
			}

			if !term.IsTerminal(int(os.Stdin.Fd())) && !yes {
				msg := "stdin não é um terminal interativo; use --yes para gravar em %s, " +
					"ou --dry-run para pré-visualizar"
				return fmt.Errorf(msg, rcPath)
			}

			if term.IsTerminal(int(os.Stdin.Fd())) && !yes {
				ok, cerr := system.Confirm(
					context.Background(),
					fmt.Sprintf("Remover o autocompletar mb-cli de %s?", rcPath),
					cmd.InOrStdin(),
					cmd.ErrOrStderr(),
				)
				if cerr != nil {
					return cerr
				}
				if !ok {
					fmt.Fprintln(cmd.ErrOrStderr(), "Operação cancelada.")
					return nil
				}
			}

			mode := os.FileMode(0o644)
			if fi, statErr := os.Stat(rcPath); statErr == nil {
				mode = fi.Mode().Perm()
			}
			if err := os.WriteFile(rcPath, []byte(newContent), mode); err != nil {
				return fmt.Errorf("gravar %s: %w", rcPath, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Autocompletar mb-cli removido de %s.\n", rcPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&shellFlag, "shell", "", fmt.Sprintf(
		"Shell alvo (%s, %s, %s, %s); por omissão deteta via SHELL",
		ShellBash, ShellZsh, ShellFish, ShellPowerShell))
	rcUsage := "Ficheiro de perfil a editar " +
		"(substitui o path por omissão do shell)"
	cmd.Flags().StringVar(&rcFile, "rc-file", "", rcUsage)
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Mostra o conteúdo final sem gravar")
	yesUsage := "Grava sem pedir confirmação (obrigatório em CI / sem TTY)"
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, yesUsage)
	return cmd
}
