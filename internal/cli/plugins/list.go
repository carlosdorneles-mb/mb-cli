package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"mb/internal/deps"
	"mb/internal/domain/plugin"
	mbplugins "mb/internal/infra/plugins"
	"mb/internal/shared/system"
)

// PluginEntry representa um plugin com todos os dados (usado no JSON)
// Este tipo é espelhado em internal/shared/system/fzf.go para o preview
type PluginEntry = system.PluginEntry

func newPluginsListCmd(d deps.Dependencies) *cobra.Command {
	var checkUpdates bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista plugins instalados",
		Long: `Lista plugins instalados com informações detalhadas.

No modo interativo (terminal), mostra interface fzf com colunas
simplificadas: PACOTE | COMANDO

Em pipes/redirecionamentos, mostra tabela completa com colunas:
PACOTE | COMANDO | DESCRIÇÃO | VERSÃO | ORIGEM | ATUALIZAR

No modo interativo, o preview aparece automaticamente no lado direito
mostrando detalhes completos incluindo origem, versão e URL.

A coluna PACOTE é o identificador usado em mb plugins remove <pacote> e
mb plugins update <pacote>. Ao instalar sem --package, o nome vem do
repositório (Git) ou do diretório (local).`,
		Example: `  # Modo interativo (terminal)
  mb plugins list
  mb plugins ls

  # Com verificação de atualizações
  mb plugins list --check-updates

  # Saída em formato JSON
  mb plugins list --json
  mb plugins list -J
  mb plugins list --json | jq '.plugins[].package'

  # Pipe com tabela completa
  mb plugins list | cat
  mb plugins list | grep local
  mb plugins list | wc -l`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pluginList, err := d.Store.ListPlugins()
			if err != nil {
				return err
			}
			sources, err := d.Store.ListPluginSources()
			if err != nil {
				return err
			}
			sort.Slice(pluginList, func(i, j int) bool {
				return pluginList[i].CommandPath < pluginList[j].CommandPath
			})

			// Build complete entries for JSON and preview
			entries := buildPluginEntries(
				pluginList,
				sources,
				d.Runtime.PluginsDir,
				checkUpdates,
				cmd,
			)

			// JSON mode: output all data and exit
			if jsonOutput {
				return outputJSON(cmd, entries)
			}

			// Pipe mode: show full columns (not a TTY)
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return outputPipeTable(cmd, entries)
			}

			// Interactive mode: show fzf with preview
			headers := []string{"PACOTE", "COMANDO"}
			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				rows = append(rows, []string{e.Package, e.Command})
			}

			_, previewErr := system.FzfTableWithPreview(
				cmd.Context(),
				headers,
				rows,
				cmd.OutOrStdout(),
				entries,
			)
			return previewErr
		},
	}

	cmd.Flags().
		BoolVar(&checkUpdates, "check-updates", false, "Verifica se há atualização disponível para cada plugin")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "J", false, "Saída em formato JSON")
	return cmd
}

func buildPluginEntries(
	pluginList []plugin.Plugin,
	sources []plugin.PluginSource,
	pluginsDir string,
	checkUpdates bool,
	cmd *cobra.Command,
) []system.PluginEntry {
	entries := make([]system.PluginEntry, 0, len(pluginList))

	for _, p := range pluginList {
		src := mbplugins.SourceForPlugin(p, sources, pluginsDir)
		pkgID := p.CommandPath
		if pkgID == "" {
			pkgID = p.CommandName
		}
		if src != nil {
			pkgID = src.InstallDir
		}

		entry := system.PluginEntry{
			Package:     pkgID,
			Command:     p.CommandPath,
			Description: p.Description,
			Version:     "-",
			Origin:      "-",
			URL:         "-",
		}

		if src != nil {
			entry.Version = src.Version
			if src.LocalPath != "" {
				entry.Origin = "local"
				entry.URL = src.LocalPath
			} else if src.GitURL != "" {
				entry.Origin = "remoto"
				entry.URL = src.GitURL
				entry.Ref = src.Ref
				entry.RefType = src.RefType
			}
		}

		if checkUpdates && src != nil && src.GitURL != "" && src.LocalPath == "" {
			dir := filepath.Join(pluginsDir, src.InstallDir)
			if mbplugins.IsGitRepo(dir) {
				if src.RefType == "tag" {
					_ = mbplugins.FetchTags(cmd.Context(), dir)
					tags, _ := mbplugins.ListLocalTags(dir)
					for _, t := range tags {
						if _, newer := mbplugins.NewerTag(src.Ref, t); newer {
							entry.UpdateAvailable = true
							break
						}
					}
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries
}

func outputJSON(cmd *cobra.Command, entries []system.PluginEntry) error {
	output := struct {
		Plugins []system.PluginEntry `json:"plugins"`
	}{
		Plugins: entries,
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputPipeTable(cmd *cobra.Command, entries []PluginEntry) error {
	headers := []string{"PACOTE", "COMANDO", "DESCRIÇÃO", "VERSÃO", "ORIGEM", "ATUALIZAR"}
	rows := make([][]string, 0, len(entries))

	for _, e := range entries {
		// Truncate description to 47 chars + "..." = 50 total
		desc := e.Description
		if len(desc) > 47 {
			desc = desc[:44] + "..."
		}

		// Update availability
		updateStr := "-"
		if e.UpdateAvailable {
			updateStr = "sim"
		}

		rows = append(rows, []string{
			e.Package,
			e.Command,
			desc,
			e.Version,
			e.Origin,
			updateStr,
		})
	}

	return system.GumTable(cmd.Context(), headers, rows, cmd.OutOrStdout())
}
