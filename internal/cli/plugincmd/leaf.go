package plugincmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	domainplugin "mb/internal/domain/plugin"
	"mb/internal/ports"
	"mb/internal/shared/pluginutil"
	"mb/internal/shared/system"
)

func applyCobraPluginFields(cmd *cobra.Command, p domainplugin.Plugin, defaultUse string) {
	if p.UseTemplate != "" {
		cmd.Use = defaultUse + " " + strings.TrimSpace(p.UseTemplate)
	} else {
		cmd.Use = defaultUse
	}
	if p.ArgsCount > 0 {
		cmd.Args = cobra.ExactArgs(p.ArgsCount)
	}
	if p.AliasesJSON != "" {
		var aliases []string
		if err := json.Unmarshal([]byte(p.AliasesJSON), &aliases); err == nil {
			cmd.Aliases = aliases
		}
	}
	if p.Example != "" {
		cmd.Example = p.Example
	}
	if p.LongDescription != "" {
		cmd.Long = p.LongDescription
	}
	if p.Deprecated != "" && cmd.RunE != nil {
		oldRunE := cmd.RunE
		deprecatedMsg := p.Deprecated
		cmdName := defaultUse
		cmd.RunE = func(c *cobra.Command, args []string) error {
			fmt.Fprintf(c.ErrOrStderr(), "Comando %q está obsoleto: %s\n", cmdName, deprecatedMsg)
			return oldRunE(c, args)
		}
	}
}

func newLeafCommand(
	use string,
	p domainplugin.Plugin,
	d deps.Dependencies,
	exec ports.ScriptExecutor,
	pluginRoot string,
	isLocal bool,
	dbgLog *system.Logger,
	globalShorts map[string]struct{},
) *cobra.Command {
	short := p.Description
	if short == "" {
		short = "Executa " + p.CommandPath
	}
	if isLocal {
		short += " (local)"
	}

	if p.FlagsJSON == "" {
		cmd := &cobra.Command{
			Use:   use,
			Short: short,
			RunE:  runEntrypointCommand(p, d, exec, pluginRoot),
		}
		applyCobraPluginFields(cmd, p, use)
		registerReadmeFlag(cmd, p.CommandPath, p.ReadmePath, dbgLog, nil, globalShorts)
		cmd.Flags().ParseErrorsAllowlist.UnknownFlags = true
		setHelpFang(cmd)
		return cmd
	}

	var flagsMap map[string]pluginutil.FlagDef
	if err := json.Unmarshal([]byte(p.FlagsJSON), &flagsMap); err != nil {
		cmd := &cobra.Command{
			Use:    use,
			Short:  short + " (config de flags inválida)",
			Hidden: p.Hidden,
		}
		return cmd
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  runFlagsOnlyCommand(p, flagsMap, d, exec, pluginRoot),
	}
	applyCobraPluginFields(cmd, p, use)

	usedShorts := make(map[string]bool)
	for sh := range globalShorts {
		usedShorts[sh] = true
	}
	registerReadmeFlag(cmd, p.CommandPath, p.ReadmePath, dbgLog, usedShorts, globalShorts)

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
					p.CommandPath,
					name,
					sh,
					"reservado pela CLI mb",
				)
			} else if usedShorts[sh] {
				pluginFlagShortDisabledDebug(
					dbgLog,
					p.CommandPath,
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
					p.CommandPath,
					name,
					nameSh,
					"reservado pela CLI mb",
				)
				cmd.Flags().Bool(name, false, usage)
			} else if usedShorts[nameSh] {
				pluginFlagShortDisabledDebug(
					dbgLog,
					p.CommandPath,
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
