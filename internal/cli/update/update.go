// Package update implements the mb update command: phased updates for plugins, optional tools
// plugin (--update-all), CLI self-update, and system packages via nested mb machine update (shell plugin).
package update

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	appupdate "mb/internal/usecase/update"
	"mb/internal/deps"
	"mb/internal/ports"
	"mb/internal/shared/system"
	"mb/internal/usecase/plugins"
)

// machineSystemUpdateCommandPath is the plugin cache command_path that owns exposing
// mb update --only-system in help (must match machine/update/manifest.yaml under category machine).
const machineSystemUpdateCommandPath = "machine/update"

func storeHasMachineSystemUpdate(store ports.PluginCLIStore) bool {
	if store == nil {
		return false
	}
	list, err := store.ListPlugins()
	if err != nil {
		return false
	}
	for _, p := range list {
		if p.CommandPath == machineSystemUpdateCommandPath {
			return true
		}
	}
	return false
}

// toolsPluginCommandPath is the plugin cache command_path for the tools aggregator
// (mb-cli-plugins/tools/manifest.yaml). --only-tools is shown only when flags_json includes update-all.
const toolsPluginCommandPath = "tools"

func storeHasToolsUpdateAll(store ports.PluginCLIStore) bool {
	if store == nil {
		return false
	}
	list, err := store.ListPlugins()
	if err != nil {
		return false
	}
	for _, p := range list {
		if p.CommandPath != toolsPluginCommandPath || p.FlagsJSON == "" {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(p.FlagsJSON), &m); err != nil {
			continue
		}
		if _, ok := m["update-all"]; ok {
			return true
		}
	}
	return false
}

// NewUpdateCmd builds the root "mb update" cobra command.
func NewUpdateCmd(upSvc *plugins.UpdateService, d deps.Dependencies) *cobra.Command {
	var onlyPlugins, onlyCLI, onlySystem, onlyTools, checkOnly, jsonOut bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza o CLI, plugins e o sistema operacional (quando possível)",
		Long: `Atualiza, em sequência, os plugins instalados, opcionalmente o agregador mb tools --update-all (se existir no cache), o binário do MB CLI (conforme config) e, sem nenhum --only-*, a fase de sistema via mb machine update (plugin shell: brew/mas no macOS; apt/flatpak/snap no Linux quando disponíveis).

Sem flags, executa todas as fases habilitadas. Use --only-plugins e/ou --only-cli para escolher fases; com o agregador tools no cache e a flag --update-all no manifest (mb plugins sync), também --only-tools; com o plugin machine/update no cache, também --only-system (delega em mb machine update). Pode combinar várias flags --only-* (ex.: --only-plugins --only-cli).

--check-only só pode ser usado juntamente com --only-cli (verifica release do binário sem baixar). Com --only-cli --check-only, --json imprime no stdout um objeto JSON com versão local, última release e se há atualização (sem texto legível).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			log := system.NewLogger(
				d.Runtime != nil && d.Runtime.Quiet,
				d.Runtime != nil && d.Runtime.Verbose,
				cmd.ErrOrStderr(),
			)
			return appupdate.Run(ctx, cmd, d, log, appupdate.Options{
				OnlyPlugins: onlyPlugins,
				OnlyCLI:     onlyCLI,
				OnlySystem:  onlySystem,
				OnlyTools:   onlyTools,
				CheckOnly:   checkOnly,
				JSON:        jsonOut,
				RunAllGitPlugins: func(ctx context.Context) error {
					return upSvc.Update(ctx, plugins.UpdateRequest{}, log)
				},
			})
		},
	}
	cmd.Flags().
		BoolVar(&onlyPlugins, "only-plugins", false, "Atualiza os plugins instalados. Combinável com outros --only-*; sem nenhum --only-*, corre todas as fases.")
	if storeHasToolsUpdateAll(d.Store) {
		cmd.Flags().
			BoolVar(&onlyTools, "only-tools", false, "Atualiza todas as ferramentas instaladas. Combinável com outros --only-*; sem nenhum --only-*, corre todas as fases.")
	}
	cmd.Flags().
		BoolVar(&onlyCLI, "only-cli", false, "Atualiza o binário do MB CLI para a versão estável. Combinável com outros --only-*; sem nenhum --only-*, corre todas as fases.")
	if storeHasMachineSystemUpdate(d.Store) {
		cmd.Flags().
			BoolVar(&onlySystem, "only-system", false, "Atualiza pacotes do sistema (Homebrew/mas ou apt/flatpak/snap quando disponíveis). Combinável com outros --only-*; sem nenhum --only-*, corre todas as fases.")
	}
	cmd.Flags().
		BoolVar(&checkOnly, "check-only", false, "Só com --only-cli: verifica release do binário sem baixar; saída 2 se houver")
	cmd.Flags().
		BoolVar(&jsonOut, "json", false, "Só com --only-cli --check-only: imprime JSON (localVersion, remoteVersion, updateAvailable) no stdout")
	return cmd
}
