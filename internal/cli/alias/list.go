package alias

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/system"
)

func vaultCellMBRun(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(nenhum)"
	}
	return v
}

func truncateForTable(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

type aliasListJSONRow struct {
	Name     string   `json:"name"`
	EnvVault string   `json:"envVault"`
	Command  []string `json:"command"`
}

func newListCmd(d deps.Dependencies) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista aliases (fzf e preview no TTY; tabela em pipe; --json para jq)",
		Long: `Lista os aliases guardados.

No terminal interativo (stdout é TTY), mostra fzf com colunas ALIAS | VAULT e um painel
de preview à direita (vault para mb run e comando completo). O preview interativo requer
fzf, gum e jq no PATH.

Em pipe ou redirecção (ex.: mb alias list | cat, | grep, | wc), mostra uma tabela com
gum (ALIAS, VAULT, COMANDO truncado).

Com --json emite sempre um array JSON (ordenado por nome), adequado para jq, sem depender de TTY.`,
		Example: `  # Modo interativo (terminal)
  mb alias list

  # Tabela em pipe (não abre fzf)
  mb alias list | cat
  mb alias list | grep dev

  # JSON para jq
  mb alias list --json | jq '.[] | select(.envVault == "staging")'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			f, err := alib.Load(alib.FilePath(d.Runtime.ConfigDir))
			if err != nil {
				return err
			}
			names := alib.SortedNames(f)
			if len(names) == 0 {
				if asJSON {
					out := cmd.OutOrStdout()
					b, mErr := json.MarshalIndent([]aliasListJSONRow{}, "", "  ")
					if mErr != nil {
						return mErr
					}
					_, err = fmt.Fprintln(out, string(b))
					return err
				}
				msg := "Ainda não há aliases registrados. Crie um com: mb alias set <nome> -- <comando>"
				fmt.Fprintln(cmd.OutOrStdout(), msg)
				return nil
			}

			if asJSON {
				rows := make([]aliasListJSONRow, 0, len(names))
				for _, name := range names {
					e := f.Aliases[name]
					rows = append(rows, aliasListJSONRow{
						Name:     name,
						EnvVault: e.EnvVault,
						Command:  append([]string(nil), e.Command...),
					})
				}
				b, err := json.MarshalIndent(rows, "", "  ")
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return err
			}

			entries := make([]system.AliasEntry, 0, len(names))
			fzfRows := make([][]string, 0, len(names))
			for _, name := range names {
				e := f.Aliases[name]
				cmdLine := strings.Join(e.Command, " ")
				entries = append(entries, system.AliasEntry{
					Name:     name,
					EnvVault: e.EnvVault,
					Command:  cmdLine,
				})
				fzfRows = append(fzfRows, []string{name, vaultCellMBRun(e.EnvVault)})
			}

			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return outputAliasPipeTable(ctx, cmd, names, f)
			}

			headers := []string{"ALIAS", "VAULT"}
			_, err = system.FzfTableWithPreviewForAliases(
				ctx, headers, fzfRows, cmd.OutOrStdout(), entries,
			)
			return err
		},
	}

	cmd.Flags().BoolVar(
		&asJSON, "json", false,
		"Emite aliases como array JSON (name, envVault, command[])",
	)
	return cmd
}

func outputAliasPipeTable(
	ctx context.Context,
	cmd *cobra.Command,
	names []string,
	f *alib.File,
) error {
	const maxCmd = 60
	headers := []string{"ALIAS", "VAULT", "COMANDO"}
	rows := make([][]string, 0, len(names))
	for _, name := range names {
		e := f.Aliases[name]
		cmdLine := strings.Join(e.Command, " ")
		rows = append(rows, []string{
			name,
			vaultCellMBRun(e.EnvVault),
			truncateForTable(cmdLine, maxCmd),
		})
	}
	return system.GumTable(ctx, headers, rows, cmd.OutOrStdout())
}
