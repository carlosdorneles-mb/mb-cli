package completion

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/shared/system"
)

// NewInstallCmd regista `completion install` no comando completion pai.
func NewInstallCmd(root *cobra.Command) *cobra.Command {
	var (
		shellFlag string
		rcFile    string
		dryRun    bool
		yes       bool
		noDesc    bool
	)

	cmd := &cobra.Command{
		Use: "install",
		// Short/Long em português são aplicados em root.customizeCompletionPT.
		Short: "Install shell completion into your profile file",
		Long: `Detects SHELL or uses --shell, generates the same script as "mb completion <shell>",
and writes an idempotent block to the shell profile (.bashrc, .zshrc, fish config.fish,
or PowerShell profile) or to --rc-file.

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

			includeDesc := !noDesc
			rcPath, err := ProfilePath(shell, rcFile)
			if err != nil {
				return err
			}

			var script bytes.Buffer
			if err := WriteCompletionScript(root, shell, includeDesc, &script); err != nil {
				return fmt.Errorf("gerar completion: %w", err)
			}
			marked := AppendMarkers(script.String())

			existing, err := os.ReadFile(rcPath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("ler %s: %w", rcPath, err)
			}
			newContent := MergeCompletionBlock(string(existing), marked, BlockBegin, BlockEnd)

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
					fmt.Sprintf("Adicionar ou atualizar o autocompletar em %s?", rcPath),
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

			if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
				return fmt.Errorf("criar diretório: %w", err)
			}

			mode := os.FileMode(0o644)
			if fi, statErr := os.Stat(rcPath); statErr == nil {
				mode = fi.Mode().Perm()
			}
			if err := os.WriteFile(rcPath, []byte(newContent), mode); err != nil {
				return fmt.Errorf("gravar %s: %w", rcPath, err)
			}

			done := "Autocompletar instalado em %s. " +
				"Abra um novo shell ou faça source desse ficheiro.\n"
			fmt.Fprintf(cmd.OutOrStdout(), done, rcPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&shellFlag, "shell", "", fmt.Sprintf(
		"Shell alvo (%s, %s, %s, %s); por omissão deteta via SHELL",
		ShellBash, ShellZsh, ShellFish, ShellPowerShell))
	rcUsage := "Ficheiro de perfil a gravar " +
		"(substitui o path por omissão do shell)"
	cmd.Flags().StringVar(&rcFile, "rc-file", "", rcUsage)
	dryUsage := "Mostra o ficheiro e o conteúdo final sem gravar"
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, dryUsage)
	yesUsage := "Grava sem pedir confirmação (obrigatório em CI / sem TTY)"
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, yesUsage)
	noDescUsage := "Gera completion sem descrições (como mb completion <shell> --no-descriptions)"
	cmd.Flags().BoolVar(&noDesc, "no-descriptions", false, noDescUsage)
	return cmd
}
