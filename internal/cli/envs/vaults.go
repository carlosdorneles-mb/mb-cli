package envs

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	appenvs "mb/internal/usecase/envs"
)

func newVaultsCmd(d deps.Dependencies) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "vaults",
		Short: "Lista vaults de env, caminho do ficheiro e número de variáveis",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := appenvs.CollectVaultRows(envPaths(d))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				b, mErr := json.Marshal(rows)
				if mErr != nil {
					return mErr
				}
				_, err = fmt.Fprintln(out, string(b))
				return err
			}
			table := make([][]string, len(rows))
			for i, r := range rows {
				table[i] = []string{r.Vault, r.Path, strconv.Itoa(r.EnvCount)}
			}
			headers := []string{"VAULT", "ARQUIVO", "ENVS"}
			return system.GumTable(cmd.Context(), headers, table, out)
		},
	}
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite JSON [{\"vault\",\"path\",\"env_count\"},...]")
	cmd.GroupID = "commands"
	return cmd
}
