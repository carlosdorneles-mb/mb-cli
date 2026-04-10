package envs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	appenvs "mb/internal/usecase/envs"
)

func newVaultsCmd(d deps.Dependencies) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "vaults",
		Short: "Lista vaults de env e o caminho do ficheiro de cada um",
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
				table[i] = []string{r.Vault, r.Path}
			}
			headers := []string{"VAULT", "ARQUIVO"}
			return system.GumTable(cmd.Context(), headers, table, out)
		},
	}
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite JSON [{\"vault\":\"...\",\"path\":\"...\"},...]")
	cmd.GroupID = "commands"
	return cmd
}
