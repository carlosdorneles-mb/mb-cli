package envs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	appenvs "mb/internal/usecase/envs"
	"mb/internal/deps"
	"mb/internal/shared/system"
)

func newGroupsCmd(d deps.Dependencies) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "groups",
		Aliases: []string{"group"},
		Short:   "Lista grupos de env e o caminho do ficheiro de cada um",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := appenvs.CollectEnvGroupRows(envPaths(d))
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
				table[i] = []string{r.Group, r.Path}
			}
			headers := []string{"GRUPO", "ARQUIVO"}
			return system.GumTable(cmd.Context(), headers, table, out)
		},
	}
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite JSON [{\"group\":\"...\",\"path\":\"...\"},...]")
	cmd.GroupID = "commands"
	return cmd
}
