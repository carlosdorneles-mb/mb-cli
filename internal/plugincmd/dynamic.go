package plugincmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/deps"
	"mb/internal/plugins"
	"mb/internal/shared/env"
	"mb/internal/shared/safepath"
	"mb/internal/shared/system"
	"mb/internal/shared/ui"
)

// Attach registers plugin commands from the cache under root.
func Attach(root *cobra.Command, d deps.Dependencies) {
	pluginList, err := d.Store.ListPlugins()
	if err != nil {
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderError(err.Error()))
		fmt.Fprintln(
			root.ErrOrStderr(),
			ui.RenderInfo("Execute `mb plugins sync` para reconstruir o cache de plugins."),
		)
		return
	}
	if len(pluginList) == 0 {
		fmt.Fprintln(
			root.ErrOrStderr(),
			ui.RenderInfo("Não há plugins em cache. Execute `mb plugins sync` primeiro."),
		)
		return
	}

	categoryList, _ := d.Store.ListCategories()
	categoriesByPath := make(map[string]cache.Category)
	for _, c := range categoryList {
		categoriesByPath[c.Path] = c
	}

	sources, _ := d.Store.ListPluginSources()
	sourceByDir := make(map[string]*cache.PluginSource)
	for i := range sources {
		sourceByDir[sources[i].InstallDir] = &sources[i]
	}

	sort.Slice(pluginList, func(i, j int) bool {
		return pluginList[i].CommandPath < pluginList[j].CommandPath
	})

	helpGroups, _ := d.Store.ListPluginHelpGroups()
	registeredHelp := make(map[string]struct{})
	for _, hg := range helpGroups {
		root.AddGroup(&cobra.Group{ID: hg.GroupID, Title: hg.Title})
		registeredHelp[hg.GroupID] = struct{}{}
	}
	// Inconsistências de group_id no cache: debug via gum log (mesmo pipeline que mb plugins sync -v).
	dbgLog := system.NewLogger(d.Runtime.Quiet, d.Runtime.Verbose, root.ErrOrStderr())

	// Cobra requires every child GroupID to exist on the direct parent. Category commands
	// get this via ensureNestedPluginHelpGroups; a manifest leaf (e.g. tools) later used
	// as parent for tools/bruno must receive the same groups before AddCommand.
	ensureNestedPluginHelpGroups := func(cmd *cobra.Command) {
		if cmd == nil {
			return
		}
		if !cmd.ContainsGroup("commands") {
			cmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})
		}
		for _, hg := range helpGroups {
			if hg.GroupID == "" || cmd.ContainsGroup(hg.GroupID) {
				continue
			}
			cmd.AddGroup(&cobra.Group{ID: hg.GroupID, Title: hg.Title})
		}
	}

	type node struct {
		cmd     *cobra.Command
		plugins []cache.Plugin
	}
	byPath := map[string]*node{}

	for _, plugin := range pluginList {
		segments := strings.Split(plugin.CommandPath, "/")
		if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
			segments = []string{plugin.CommandName}
		}

		var parent = root
		for i := 0; i < len(segments)-1; i++ {
			seg := segments[i]
			pathSoFar := strings.Join(segments[:i+1], "/")
			if byPath[pathSoFar] == nil {
				cat := categoriesByPath[pathSoFar]
				short := seg + " COMANDOS"
				if cat.Description != "" {
					short = cat.Description
				}
				categoryCmd := &cobra.Command{
					Use:   seg,
					Short: short,
				}
				if parent == root {
					categoryCmd.GroupID = "plugin_commands"
				} else if cat.GroupID != "" {
					if _, ok := registeredHelp[cat.GroupID]; ok {
						categoryCmd.GroupID = cat.GroupID
					} else {
						_ = dbgLog.Debug(
							context.Background(),
							"plugin help: category_path=%q group_id=%q não registado no cache; usando COMANDOS",
							pathSoFar,
							cat.GroupID,
						)
						categoryCmd.GroupID = "commands"
					}
				} else {
					categoryCmd.GroupID = "commands"
				}
				if cat.ReadmePath != "" {
					categoryCmd.Flags().BoolP("readme", "r", false, readmeFlagDesc)
					categoryCmd.RunE = func(cmd *cobra.Command, _ []string) error {
						if cmd.Flags().Lookup("readme").Changed {
							return runReadmeWithGlow(cat.ReadmePath)
						}
						cmd.Help()
						return nil
					}
				}
				setHelpFang(categoryCmd)
				categoryCmd.Hidden = cat.Hidden
				if parent != root {
					ensureNestedPluginHelpGroups(parent)
				}
				ensureNestedPluginHelpGroups(categoryCmd)
				parent.AddCommand(categoryCmd)
				byPath[pathSoFar] = &node{cmd: categoryCmd}
			}
			parent = byPath[pathSoFar].cmd
		}

		pathSoFar := plugin.CommandPath
		if byPath[pathSoFar] != nil {
			continue
		}
		src := plugins.SourceForPlugin(plugin, sources, d.Runtime.PluginsDir)
		pluginRoot := plugin.PluginDir
		if pluginRoot == "" {
			installDir := plugins.FirstPathSegment(plugin.CommandPath)
			s := sourceByDir[installDir]
			pluginRoot = filepath.Join(d.Runtime.PluginsDir, installDir)
			if s != nil && s.LocalPath != "" {
				pluginRoot = s.LocalPath
			}
		}
		isLocal := src != nil && src.LocalPath != ""
		leafCmd := newLeafCommand(plugin.CommandName, plugin, d, pluginRoot, isLocal)
		leafCmd.Hidden = plugin.Hidden
		if parent == root {
			leafCmd.GroupID = "plugin_commands"
		} else if plugin.GroupID != "" {
			if _, ok := registeredHelp[plugin.GroupID]; ok {
				leafCmd.GroupID = plugin.GroupID
			} else {
				_ = dbgLog.Debug(
					context.Background(),
					"plugin help: command_path=%q group_id=%q não registado no cache; usando COMANDOS",
					plugin.CommandPath,
					plugin.GroupID,
				)
				leafCmd.GroupID = "commands"
			}
		} else {
			leafCmd.GroupID = "commands"
		}
		if parent != root {
			ensureNestedPluginHelpGroups(parent)
		}
		parent.AddCommand(leafCmd)
		byPath[pathSoFar] = &node{cmd: leafCmd, plugins: []cache.Plugin{plugin}}
	}
}

func applyCobraPluginFields(cmd *cobra.Command, plugin cache.Plugin, defaultUse string) {
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
			fmt.Fprintf(c.ErrOrStderr(), "Commando %q está obsoleto: %s\n", cmdName, deprecatedMsg)
			return oldRunE(c, args)
		}
	}
}

func newLeafCommand(
	use string,
	plugin cache.Plugin,
	d deps.Dependencies,
	pluginRoot string,
	isLocal bool,
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
		if plugin.ReadmePath != "" {
			cmd.Flags().BoolP("readme", "r", false, readmeFlagDesc)
		}
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
	for name, def := range flagsMap {
		usage := def.Description
		useShort := def.Short != "" && len([]rune(def.Short)) == 1 && !usedShorts[def.Short]
		if useShort {
			usedShorts[def.Short] = true
		}
		switch {
		case useShort:
			cmd.Flags().BoolP(name, def.Short, false, usage)
		case def.Type == "long":
			cmd.Flags().Bool(name, false, usage)
		case def.Type == "short" && len(name) == 1:
			cmd.Flags().BoolP(name, name, false, usage)
		case def.Type == "short":
			cmd.Flags().Bool(name, false, usage)
		}
	}
	if plugin.ReadmePath != "" {
		cmd.Flags().BoolP("readme", "r", false, readmeFlagDesc)
	}
	setHelpFang(cmd)
	return cmd
}

func parseRootVerbosityFlags(cmd *cobra.Command, args []string) []string {
	root := cmd.Root()
	if root == nil {
		return args
	}
	fs := root.PersistentFlags()
	_ = fs.Parse(args)
	return fs.Args()
}

func runEntrypointCommand(
	plugin cache.Plugin,
	d deps.Dependencies,
	pluginRoot string,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		argsToPass := cmd.Flags().Args()
		cliValues, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues)
		if err != nil {
			return err
		}
		fileValues, err := buildEnvFileValues(d.Runtime)
		if err != nil {
			return err
		}
		if err := mergeManifestEnvIntoFileValues(fileValues, plugin, d.Runtime); err != nil {
			return err
		}
		merged := env.Merge(os.Environ(), fileValues, cliValues)
		merged = ui.PrependGumThemeDefaults(merged)
		merged = appendVerbosityEnv(merged, d.Runtime)
		merged = appendShellHelpersEnv(merged, d.Runtime.ConfigDir)
		ctx := cmd.Context()
		if d.Runtime.PluginTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
			defer cancel()
		}
		return d.Executor.Run(ctx, plugin, argsToPass, merged, pluginRoot)
	}
}

func runFlagsOnlyCommand(
	plugin cache.Plugin,
	flagsMap map[string]plugins.FlagDef,
	d deps.Dependencies,
	pluginRoot string,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		_ = parseRootVerbosityFlags(cmd, args)
		var chosenFlag string
		var chosenEntrypoint string
		for name, def := range flagsMap {
			changed := false
			if f := cmd.Flags().Lookup(name); f != nil {
				changed = f.Changed
			}
			if changed {
				chosenFlag = name
				chosenEntrypoint = def.Entrypoint
				break
			}
		}

		if chosenFlag == "" || chosenEntrypoint == "" {
			if plugin.ExecPath != "" {
				argsToPass := cmd.Flags().Args()
				cliValues, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues)
				if err != nil {
					return err
				}
				fileValues, err := buildEnvFileValues(d.Runtime)
				if err != nil {
					return err
				}
				if err := mergeManifestEnvIntoFileValues(
					fileValues,
					plugin,
					d.Runtime,
				); err != nil {
					return err
				}
				merged := env.Merge(os.Environ(), fileValues, cliValues)
				merged = ui.PrependGumThemeDefaults(merged)
				merged = appendVerbosityEnv(merged, d.Runtime)
				merged = appendShellHelpersEnv(merged, d.Runtime.ConfigDir)
				ctx := cmd.Context()
				if d.Runtime.PluginTimeout > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
					defer cancel()
				}
				return d.Executor.Run(ctx, plugin, argsToPass, merged, pluginRoot)
			}
			cmd.Help()
			return nil
		}

		baseDir := plugin.PluginDir
		if baseDir == "" {
			segments := strings.Split(plugin.CommandPath, "/")
			if len(segments) > 1 {
				baseDir = filepath.Join(pluginRoot, filepath.Join(segments[1:]...))
			} else {
				baseDir = pluginRoot
			}
		}
		execPath := filepath.Join(baseDir, chosenEntrypoint)
		if err := safepath.ValidateUnderDir(execPath, baseDir); err != nil {
			return fmt.Errorf("flag entrypoint fora do diretório do plugin: %w", err)
		}
		pluginType := plugins.PluginTypeFromEntrypoint(chosenEntrypoint)
		syntheticPlugin := cache.Plugin{
			CommandPath:  plugin.CommandPath,
			CommandName:  plugin.CommandName,
			ExecPath:     execPath,
			PluginType:   pluginType,
			ConfigHash:   plugin.ConfigHash,
			PluginDir:    baseDir,
			EnvFilesJSON: plugin.EnvFilesJSON,
		}

		cliValues, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues)
		if err != nil {
			return err
		}
		fileValues, err := buildEnvFileValues(d.Runtime)
		if err != nil {
			return err
		}
		if err := mergeManifestEnvIntoFileValues(
			fileValues,
			syntheticPlugin,
			d.Runtime,
		); err != nil {
			return err
		}
		merged := env.Merge(os.Environ(), fileValues, cliValues)
		merged = ui.PrependGumThemeDefaults(merged)
		merged = appendVerbosityEnv(merged, d.Runtime)
		merged = appendShellHelpersEnv(merged, d.Runtime.ConfigDir)
		ctx := cmd.Context()
		if d.Runtime.PluginTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
			defer cancel()
		}
		return d.Executor.Run(ctx, syntheticPlugin, cmd.Flags().Args(), merged, pluginRoot)
	}
}

func appendShellHelpersEnv(merged []string, configDir string) []string {
	path := filepath.Join(configDir, "lib", "shell")
	return append(merged, "MB_HELPERS_PATH="+path)
}

func appendVerbosityEnv(merged []string, rt *deps.RuntimeConfig) []string {
	if rt == nil {
		return merged
	}
	if rt.Quiet {
		merged = append(merged, "MB_QUIET=1")
	}
	if rt.Verbose {
		merged = append(merged, "MB_VERBOSE=1")
	}
	return merged
}

const readmeFlagDesc = "Visualizar documentação do commando"

func setHelpFang(c *cobra.Command) {
	c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if root := cmd.Root(); root != nil {
			root.HelpFunc()(cmd, args)
		}
	})
}

func runReadmeWithGlow(path string) error {
	if path == "" {
		return errors.New("este commando não possui documentação (README) disponível")
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("documentação não encontrada para este commando")
		}
		return fmt.Errorf("não foi possível abrir a documentação: %w", err)
	}
	return system.RenderMarkdown(context.Background(), path)
}

func buildEnvFileValues(rt *deps.RuntimeConfig) (map[string]string, error) {
	return deps.BuildEnvFileValues(rt)
}

func mergeManifestEnvIntoFileValues(
	fileValues map[string]string,
	plugin cache.Plugin,
	rt *deps.RuntimeConfig,
) error {
	group := plugins.ManifestEnvGroupDefault
	if rt != nil && strings.TrimSpace(rt.EnvGroup) != "" {
		group = strings.TrimSpace(rt.EnvGroup)
	}
	extra, err := plugins.MergeManifestEnvFiles(plugin.PluginDir, plugin.EnvFilesJSON, group)
	if err != nil {
		return err
	}
	for k, v := range extra {
		fileValues[k] = v
	}
	return nil
}
