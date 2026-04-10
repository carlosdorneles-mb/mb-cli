package envs

import (
	"github.com/spf13/cobra"

	"mb/internal/usecase/envs"
)

func newListCmd(svc *envs.ListService) *cobra.Command {
	var listVault string
	var asJSON, asText, showSecrets bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis do vault padrão ou de um vault específico",
		RunE: func(cmd *cobra.Command, _ []string) error {
			format := envs.FormatTable
			if asJSON {
				format = envs.FormatJSON
			} else if asText {
				format = envs.FormatText
			}
			rows, err := svc.List(cmd.Context(), envs.ListRequest{
				Vault:       listVault,
				ShowSecrets: showSecrets,
			})
			if err != nil {
				return err
			}
			return envs.FormatRows(cmd.Context(), cmd.OutOrStdout(), rows, format)
		},
	}
	cmd.Flags().
		StringVar(&listVault, "vault", "", "Filtra por vault: nome em ~/.config/mb, ou project / project/<nome> só em mbcli.yaml")
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
