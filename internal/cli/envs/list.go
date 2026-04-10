package envs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	appenvs "mb/internal/usecase/envs"
	"mb/internal/deps"
	"mb/internal/shared/system"
)

func newListCmd(d deps.Dependencies) *cobra.Command {
	var listGroup string
	var asJSON, asText, showSecrets bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis padrão ou de um grupo específico",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := appenvs.CollectListRows(
				d.SecretStore,
				d.OnePassword,
				envPaths(d),
				listGroup,
				showSecrets,
			)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch {
			case asJSON:
				obj := make(map[string]string, len(rows))
				for _, r := range rows {
					obj[r.Key] = r.Value
				}
				b, mErr := json.Marshal(obj)
				if mErr != nil {
					return mErr
				}
				_, err = fmt.Fprintln(out, string(b))
				return err
			case asText:
				for _, r := range rows {
					if _, err = fmt.Fprintf(out, "%s=%s\n", r.Key, r.Value); err != nil {
						return err
					}
				}
				return nil
			default:
				table := make([][]string, len(rows))
				for i, r := range rows {
					table[i] = []string{r.Key + "=" + r.Value, r.Group, r.Storage}
				}
				headers := []string{"VAR", "GRUPO", "ARMAZENAMENTO"}
				return system.GumTable(cmd.Context(), headers, table, out)
			}
		},
	}
	cmd.Flags().StringVar(&listGroup, "group", "", "Lista apenas variáveis do grupo informado")
	cmd.Flags().
		BoolVar(&showSecrets, "show-secrets", false, "Mostra o valor real das variáveis guardadas no keyring (por defeito mostram ***)")
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite variáveis como objeto JSON {\"CHAVE\":\"valor\",...}")
	cmd.Flags().
		BoolVarP(&asText, "text", "T", false, "Emite somente key=value por linha (sem grupo)")
	cmd.MarkFlagsMutuallyExclusive("json", "text")
	cmd.GroupID = "commands"
	return cmd
}
