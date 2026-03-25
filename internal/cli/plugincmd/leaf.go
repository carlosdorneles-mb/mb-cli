package plugincmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/system"
)

func applyCobraPluginFields(cmd *cobra.Command, plugin sqlite.Plugin, defaultUse string) {
	if plugin.UseTemplate != "" {
		cmd.Use = defaultUse + " " + strings.TrimSpace(plugin.UseTemplate)
	} else {
		cmd.Use = defaultUse
	}
	if plugin.ArgsCount > 0 {
		cmd.Args = cobra.ExactArgs(plugin.ArgsCount)
	}
	if plugin.AliasesJSON != "" {
		var aliases []string
		if err := json.Unmarshal([]byte(plugin.AliasesJSON), &aliases); err == nil {
			cmd.Aliases = aliases
		}
	}
	if plugin.Example != "" {
		cmd.Example = plugin.Example
	}
	if plugin.LongDescription != "" {
		cmd.Long = plugin.LongDescription
	}
	if plugin.Deprecated != "" && cmd.RunE != nil {
		oldRunE := cmd.RunE
		deprecatedMsg := plugin.Deprecated
		cmdName := defaultUse
		cmd.RunE = func(c *cobra.Command, args []string) error {
			fmt.Fprintf(c.ErrOrStderr(), "Comando %q está obsoleto: %s\n", cmdName, deprecatedMsg)
			return oldRunE(c, args)
		}
	}
}

func newLeafCommand(
	use string,
	plugin sqlite.Plugin,
	d deps.Dependencies,
	pluginRoot string,
	isLocal bool,
	dbgLog *system.Logger,
	globalShorts map[string]struct{},
) *cobra.Command {
	short := plugin.Description
	if short == "" {
		short = "Executa " + plugin.CommandPath
	}
	if isLocal {
		short += " (local)"
	}

	if plugin.FlagsJSON == "" {
		cmd := &cobra.Command{
			Use:   use,
			Short: short,
			RunE:  runEntrypointCommand(plugin, d, pluginRoot),
		}
		applyCobraPluginFields(cmd, plugin, use)
		registerReadmeFlag(cmd, plugin.CommandPath, plugin.ReadmePath, dbgLog, nil, globalShorts)
		cmd.Flags().ParseErrorsAllowlist.UnknownFlags = true
		setHelpFang(cmd)
		return cmd
	}

	var flagsMap map[string]plugins.FlagDef
	if err := json.Unmarshal([]byte(plugin.FlagsJSON), &flagsMap); err != nil {
		cmd := &cobra.Command{
			Use:    use,
			Short:  short + " (config de flags inválida)",
			Hidden: plugin.Hidden,
		}
		return cmd
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  runFlagsOnlyCommand(plugin, flagsMap, d, pluginRoot),
	}
	applyCobraPluginFields(cmd, plugin, use)

	usedShorts := make(map[string]bool)
	for sh := range globalShorts {
		usedShorts[sh] = true
	}
	registerReadmeFlag(cmd, plugin.CommandPath, plugin.ReadmePath, dbgLog, usedShorts, globalShorts)

	for name, def := range flagsMap {
		usage := def.Description
		sh := ""
		if def.Short != "" && len([]rune(def.Short)) == 1 {
			sh = def.Short
		}
		useShort := false
		if sh != "" {
			if globalShorthandReserved(globalShorts, sh) {
				pluginFlagShortDisabledDebug(
					dbgLog,
					plugin.CommandPath,
					name,
					sh,
					"reservado pela CLI mb",
				)
			} else if usedShorts[sh] {
				pluginFlagShortDisabledDebug(
					dbgLog,
					plugin.CommandPath,
					name,
					sh,
					"já usado neste comando",
				)
			} else {
				useShort = true
				usedShorts[sh] = true
			}
		}
		switch {
		case useShort:
			cmd.Flags().BoolP(name, sh, false, usage)
		case def.Type == "long":
			cmd.Flags().Bool(name, false, usage)
		case def.Type == "short" && len(name) == 1:
			nameSh := name
			if globalShorthandReserved(globalShorts, nameSh) {
				pluginFlagShortDisabledDebug(
					dbgLog,
					plugin.CommandPath,
					name,
					nameSh,
					"reservado pela CLI mb",
				)
				cmd.Flags().Bool(name, false, usage)
			} else if usedShorts[nameSh] {
				pluginFlagShortDisabledDebug(
					dbgLog,
					plugin.CommandPath,
					name,
					nameSh,
					"já usado neste comando",
				)
				cmd.Flags().Bool(name, false, usage)
			} else {
				usedShorts[nameSh] = true
				cmd.Flags().BoolP(name, nameSh, false, usage)
			}
		case def.Type == "short":
			cmd.Flags().Bool(name, false, usage)
		}
	}
	setHelpFang(cmd)
	return cmd
}
