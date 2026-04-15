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
	"mb/internal/shared/system"
)

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
	Name      string   `json:"name"`
	EnvVault  string   `json:"envVault"`
	Command   []string `json:"command"`
	Source    string   `json:"source"`
	MbcliPath string   `json:"mbcliPath"`
}

func newListCmd(d deps.Dependencies) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista aliases (fzf e preview no TTY; tabela em pipe; --json para jq)",
		Long: `Lista os aliases salvos.

Inclui aliases em ~/.config/mb/aliases.yaml e aliases de repositório em mbcli.yaml (quando
resolvido por MBCLI_YAML_PATH / MBCLI_PROJECT_ROOT); em conflito de nome, vale a definição
do repositório (a mesma precedência do mb run).

No terminal interativo (stdout é TTY), mostra fzf com colunas ALIAS | VAULT | FONTE e um painel
de preview à direita (vault para mb run e comando completo). O preview interativo requer
fzf, gum e jq no PATH.

Em pipe ou redirecionamento (ex.: mb alias list | cat, | grep, | wc), mostra uma tabela com
gum (ALIAS, VAULT, COMANDO truncado, FONTE).

Com --json emite sempre um array JSON (ordenado por nome), adequado para jq, sem depender de TTY.`,
		Example: `  # Modo interativo (terminal)
  mb alias list

  # Tabela em pipe (não abre fzf)
  mb alias list | cat
  mb alias list | grep dev

  # JSON para jq
  mb alias list --json | jq '.[] | select(.source == "project")'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			rowsMerged, err := loadMergedAliasRows(d.Runtime.ConfigDir)
			if err != nil {
				return err
			}
			if len(rowsMerged) == 0 {
				if asJSON {
					out := cmd.OutOrStdout()
					b, mErr := json.MarshalIndent([]aliasListJSONRow{}, "", "  ")
					if mErr != nil {
						return mErr
					}
					_, err = fmt.Fprintln(out, string(b))
					return err
				}
				msg := "Ainda não há aliases salvos. Crie um com: mb alias set <nome> -- <comando>"
				fmt.Fprintln(cmd.OutOrStdout(), msg)
				return nil
			}

			if asJSON {
				jsonRows := make([]aliasListJSONRow, 0, len(rowsMerged))
				for _, row := range rowsMerged {
					jsonRows = append(jsonRows, aliasListJSONRow{
						Name:      row.Name,
						EnvVault:  aliasListVaultDisplay(row.Source, row.EnvVault),
						Command:   append([]string(nil), row.Command...),
						Source:    row.Source,
						MbcliPath: row.MbcliPath,
					})
				}
				b, err := json.MarshalIndent(jsonRows, "", "  ")
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return err
			}

			entries := make([]system.AliasEntry, 0, len(rowsMerged))
			fzfRows := make([][]string, 0, len(rowsMerged))
			for _, row := range rowsMerged {
				cmdLine := strings.Join(row.Command, " ")
				vDisp := aliasListVaultDisplay(row.Source, row.EnvVault)
				entries = append(entries, system.AliasEntry{
					Name:      row.Name,
					EnvVault:  vDisp,
					Command:   cmdLine,
					Source:    row.Source,
					MbcliPath: row.MbcliPath,
				})
				fzfRows = append(fzfRows, []string{
					row.Name,
					vDisp,
					sourceCell(row.Source),
				})
			}

			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return outputAliasPipeTable(ctx, cmd, rowsMerged)
			}

			headers := []string{"ALIAS", "VAULT", "FONTE"}
			_, err = system.FzfTableWithPreviewForAliases(
				ctx, headers, fzfRows, cmd.OutOrStdout(), entries,
			)
			return err
		},
	}

	cmd.Flags().BoolVar(
		&asJSON, "json", false,
		"Emite aliases como array JSON (name, envVault com rótulo project/project/n para mbcli.yaml, command[], source, mbcliPath)",
	)
	return cmd
}

func sourceCell(s string) string {
	switch s {
	case "project":
		return "repositório"
	case "config":
		return "config"
	default:
		return s
	}
}

func outputAliasPipeTable(
	ctx context.Context,
	cmd *cobra.Command,
	rowsMerged []aliasListRow,
) error {
	const maxCmd = 60
	headers := []string{"ALIAS", "VAULT", "COMANDO", "FONTE"}
	rows := make([][]string, 0, len(rowsMerged))
	for _, row := range rowsMerged {
		cmdLine := strings.Join(row.Command, " ")
		rows = append(rows, []string{
			row.Name,
			aliasListVaultDisplay(row.Source, row.EnvVault),
			truncateForTable(cmdLine, maxCmd),
			sourceCell(row.Source),
		})
	}
	return system.GumTable(ctx, headers, rows, cmd.OutOrStdout())
}
