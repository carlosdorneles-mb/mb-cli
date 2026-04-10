package envs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/system"
	appenvs "mb/internal/usecase/envs"
)

func newListCmd(d deps.Dependencies) *cobra.Command {
	var listVault string
	var asJSON, asText, showSecrets bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis do vault padrão ou de um vault específico",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := appenvs.CollectListRows(
				d.SecretStore,
				d.OnePassword,
				envPaths(d),
				listVault,
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
					table[i] = []string{r.Key + "=" + r.Value, r.Vault, r.Storage}
				}
				headers := []string{"VAR", "VAULT", "ARMAZENAMENTO"}
				return system.GumTable(cmd.Context(), headers, table, out)
			}
		},
	}
	cmd.Flags().StringVar(&listVault, "vault", "", "Lista apenas variáveis do vault informado")
	cmd.Flags().
		BoolVar(&showSecrets, "show-secrets", false, "Mostra o valor real das variáveis guardadas no keyring (por defeito mostram ***)")
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite variáveis como objeto JSON {\"CHAVE\":\"valor\",...}")
	cmd.Flags().
		BoolVarP(&asText, "text", "T", false, "Emite somente key=value por linha (sem vault)")
	cmd.MarkFlagsMutuallyExclusive("json", "text")
	cmd.GroupID = "commands"
	return cmd
}
