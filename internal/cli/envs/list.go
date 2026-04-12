package envs

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/shared/system"
	"mb/internal/usecase/envs"
)

func newListCmd(svc *envs.ListService, configDir string) *cobra.Command {
	var listVault string
	var asJSON, asText, showSecrets bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis do vault padrão ou de um vault específico",
		Long: `Lista variáveis de ambiente com informações detalhadas.

No modo interativo (terminal), mostra interface fzf com colunas
simplificadas: VAR | VAULT

Em pipes/redirecionamentos, mostra tabela completa com colunas:
VAR | VAULT | ARMAZENAMENTO

No modo interativo, o preview aparece automaticamente no lado direito
mostrando detalhes completos incluindo valor e tipo de armazenamento.`,
		Example: `  # Modo interativo (terminal)
  mb envs list
  mb envs ls
  mb envs l

  # Filtrar por vault
  mb envs list --vault staging
  mb envs list --vault project

  # Saída em formato JSON (array plano com vault)
  mb envs list --json
  mb envs list --json | jq '.[] | select(.key == "API_BASE") | .value'
  mb envs list --json | jq '.[] | select(.vault == "staging")'

  # Saída em formato texto (key=value)
  mb envs list --text
  mb envs list --text | grep API

  # Pipe com tabela completa
  mb envs list | cat
  mb envs list | grep API
  mb envs list | wc -l`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get rows
			rows, err := svc.List(cmd.Context(), envs.ListRequest{
				Vault:       listVault,
				ShowSecrets: showSecrets,
			})
			if err != nil {
				return err
			}

			// JSON/Text mode: existing behavior (suporta pipes)
			if asJSON {
				return envs.FormatJSONByVault(cmd.OutOrStdout(), rows, showSecrets, configDir)
			} else if asText {
				return envs.FormatTextPlain(cmd.OutOrStdout(), rows)
			}

			// Build entries for fzf preview
			entries := buildEnvEntries(rows, showSecrets, configDir)

			// Pipe mode: show full table when not a TTY
			// Supports: | cat, | grep, | wc, etc.
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return outputEnvPipeTable(cmd, entries)
			}

			// Interactive mode: fzf with preview (TTY only)
			headers := []string{"VAR", "VAULT"}
			fzfRows := make([][]string, 0, len(entries))
			for _, e := range entries {
				fzfRows = append(fzfRows, []string{e.Key, e.Vault})
			}

			_, err = system.FzfTableWithPreviewForEnv(
				cmd.Context(),
				headers,
				fzfRows,
				cmd.OutOrStdout(),
				entries,
			)
			return err
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

// buildEnvEntries converte ListRow para EnvEntry com tratamento de secrets
func buildEnvEntries(rows []envs.ListRow, showSecrets bool, configDir string) []system.EnvEntry {
	entries := make([]system.EnvEntry, 0, len(rows))

	for _, r := range rows {
		entry := system.EnvEntry{
			Key:     r.Key,
			Value:   r.Value,
			Vault:   r.Vault,
			Storage: r.Storage,
			Path:    envs.CalculateEnvFilePath(r.Vault, r.Storage, configDir),
		}

		// Detect if it's a secret
		entry.IsSecret = (r.Storage == envs.StorageKeyring ||
			r.Storage == envs.Storage1Password)

		// Mask secrets unless --show-secrets
		if entry.IsSecret && !showSecrets {
			entry.DisplayValue = "***"
		} else {
			entry.DisplayValue = r.Value
		}

		entries = append(entries, entry)
	}

	return entries
}

// outputEnvPipeTable exibe tabela completa para modo pipe
func outputEnvPipeTable(cmd *cobra.Command, entries []system.EnvEntry) error {
	headers := []string{"VAR", "VAULT", "ARMAZENAMENTO"}
	rows := make([][]string, 0, len(entries))

	for _, e := range entries {
		// Truncate long values if needed
		displayVal := e.DisplayValue
		if len(displayVal) > 60 {
			displayVal = displayVal[:57] + "..."
		}

		varCol := fmt.Sprintf("%s=%s", e.Key, displayVal)
		rows = append(rows, []string{
			varCol,
			e.Vault,
			e.Storage,
		})
	}

	return system.GumTable(cmd.Context(), headers, rows, cmd.OutOrStdout())
}
