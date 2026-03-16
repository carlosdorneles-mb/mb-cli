package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/commands/self"
	"mb/internal/env"
	"mb/internal/plugins"
	"mb/internal/system"
	"mb/internal/ui"
)

func AttachDynamicCommands(root *cobra.Command, deps config.Dependencies) {
	pluginList, err := deps.Store.ListPlugins()
	if err != nil {
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderError(err.Error()))
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderInfo("Execute `mb self sync` para reconstruir o cache de plugins."))
		return
	}
	if len(pluginList) == 0 {
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderInfo("Não há plugins em cache. Execute `mb self sync` primeiro."))
		return
	}

	categoryList, _ := deps.Store.ListCategories()
	categoriesByPath := make(map[string]cache.Category)
	for _, c := range categoryList {
		categoriesByPath[c.Path] = c
	}

	sources, _ := deps.Store.ListPluginSources()
	sourceByDir := make(map[string]*cache.PluginSource)
	for i := range sources {
		sourceByDir[sources[i].InstallDir] = &sources[i]
	}

	sort.Slice(pluginList, func(i, j int) bool {
		return pluginList[i].CommandPath < pluginList[j].CommandPath
	})

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

		var parent *cobra.Command = root
		for i := 0; i < len(segments)-1; i++ {
			seg := segments[i]
			pathSoFar := strings.Join(segments[:i+1], "/")
			if byPath[pathSoFar] == nil {
				cat := categoriesByPath[pathSoFar]
				short := seg + " comandos"
				if cat.Description != "" {
					short = cat.Description
				}
				categoryCmd := &cobra.Command{
					Use:   seg,
					Short: short,
				}
				if parent == root {
					categoryCmd.GroupID = "plugin_commands"
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
				parent.AddCommand(categoryCmd)
				byPath[pathSoFar] = &node{cmd: categoryCmd}
			}
			parent = byPath[pathSoFar].cmd
		}

		pathSoFar := plugin.CommandPath
		if byPath[pathSoFar] != nil {
			continue
		}
		installDir := self.FirstPathSegment(plugin.CommandPath)
		src := sourceByDir[installDir]
		pluginRoot := filepath.Join(deps.Runtime.PluginsDir, installDir)
		isLocal := false
		if src != nil && src.LocalPath != "" {
			pluginRoot = src.LocalPath
			isLocal = true
		}
		leafCmd := newLeafCommand(plugin.CommandName, plugin, deps, pluginRoot, isLocal)
		if parent == root {
			leafCmd.GroupID = "plugin_commands"
		}
		parent.AddCommand(leafCmd)
		byPath[pathSoFar] = &node{cmd: leafCmd, plugins: []cache.Plugin{plugin}}
	}
}

func newLeafCommand(use string, plugin cache.Plugin, deps config.Dependencies, pluginRoot string, isLocal bool) *cobra.Command {
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
			RunE:  runEntrypointCommand(plugin, deps),
		}
		if plugin.ReadmePath != "" {
			cmd.Flags().BoolP("readme", "r", false, readmeFlagDesc)
		} else {
			cmd.DisableFlagParsing = true
		}
		setHelpFang(cmd)
		return cmd
	}

	var flagsMap map[string]plugins.FlagDef
	if err := json.Unmarshal([]byte(plugin.FlagsJSON), &flagsMap); err != nil {
		cmd := &cobra.Command{Use: use, Short: short + " (config de flags inválida)"}
		return cmd
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  runFlagsOnlyCommand(plugin, flagsMap, deps, pluginRoot),
	}

	usedShorts := make(map[string]bool)
	for name, def := range flagsMap {
		useShort := def.Short != "" && len([]rune(def.Short)) == 1 && !usedShorts[def.Short]
		if useShort {
			usedShorts[def.Short] = true
		}
		switch {
		case useShort:
			cmd.Flags().BoolP(name, def.Short, false, "")
		case def.Type == "long":
			cmd.Flags().Bool(name, false, "")
		case def.Type == "short" && len(name) == 1:
			cmd.Flags().BoolP(name, name, false, "")
		case def.Type == "short":
			cmd.Flags().Bool(name, false, "")
		}
	}
	if plugin.ReadmePath != "" {
		cmd.Flags().BoolP("readme", "r", false, readmeFlagDesc)
	}
	setHelpFang(cmd)
	return cmd
}

// parseRootVerbosityFlags parses args against the root's persistent flags (e.g. -v, -q)
// so that deps.Runtime.Verbose/Quiet are set even when the flag appears after the subcommand
// (e.g. mb tools hello -v). It returns the remaining args not consumed by those flags.
func parseRootVerbosityFlags(cmd *cobra.Command, args []string) []string {
	root := cmd.Root()
	if root == nil {
		return args
	}
	fs := root.PersistentFlags()
	_ = fs.Parse(args)
	return fs.Args()
}

func runEntrypointCommand(plugin cache.Plugin, deps config.Dependencies) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		args = parseRootVerbosityFlags(cmd, args)
		argsToPass := args
		if cmd.Flags().Lookup("readme") != nil {
			argsToPass = cmd.Flags().Args()
		}
		cliValues, err := env.ParseInlinePairs(deps.Runtime.InlineEnvValues)
		if err != nil {
			return err
		}
		fileValues, err := buildEnvFileValues(deps.Runtime)
		if err != nil {
			return err
		}
		merged := env.Merge(os.Environ(), fileValues, cliValues)
		merged = ui.PrependGumThemeDefaults(merged)
		merged = appendVerbosityEnv(merged, deps.Runtime)
		merged = appendShellHelpersEnv(merged, deps.Runtime.ConfigDir)
		return deps.Executor.Run(cmd.Context(), plugin, argsToPass, merged)
	}
}

func runFlagsOnlyCommand(plugin cache.Plugin, flagsMap map[string]plugins.FlagDef, deps config.Dependencies, pluginRoot string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		args = parseRootVerbosityFlags(cmd, args)
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
			cmd.Help()
			return nil
		}

		segments := strings.Split(plugin.CommandPath, "/")
		var baseDir string
		if len(segments) > 1 {
			baseDir = filepath.Join(pluginRoot, filepath.Join(segments[1:]...))
		} else {
			baseDir = pluginRoot
		}
		execPath := filepath.Join(baseDir, chosenEntrypoint)
		pluginType := plugins.PluginTypeFromEntrypoint(chosenEntrypoint)
		syntheticPlugin := cache.Plugin{
			CommandPath: plugin.CommandPath,
			CommandName: plugin.CommandName,
			ExecPath:    execPath,
			PluginType:  pluginType,
			ConfigHash:  plugin.ConfigHash,
		}

		cliValues, err := env.ParseInlinePairs(deps.Runtime.InlineEnvValues)
		if err != nil {
			return err
		}
		fileValues, err := buildEnvFileValues(deps.Runtime)
		if err != nil {
			return err
		}
		merged := env.Merge(os.Environ(), fileValues, cliValues)
		merged = ui.PrependGumThemeDefaults(merged)
		merged = appendVerbosityEnv(merged, deps.Runtime)
		merged = appendShellHelpersEnv(merged, deps.Runtime.ConfigDir)
		return deps.Executor.Run(cmd.Context(), syntheticPlugin, args, merged)
	}
}

func appendShellHelpersEnv(merged []string, configDir string) []string {
	path := filepath.Join(configDir, "lib", "shell")
	return append(merged, "MB_HELPERS_PATH="+path)
}

func appendVerbosityEnv(merged []string, rt *config.RuntimeConfig) []string {
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

const readmeFlagDesc = "Visualizar documentação do comando"

func setHelpFang(c *cobra.Command) {
	c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if root := cmd.Root(); root != nil {
			root.HelpFunc()(cmd, args)
		}
	})
}

func runReadmeWithGlow(path string) error {
	if path == "" {
		return errors.New("este comando não possui documentação (README) disponível")
	}
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("documentação não encontrada para este comando")
		}
		return fmt.Errorf("não foi possível abrir a documentação: %w", err)
	}
	return system.RenderMarkdown(context.Background(), path)
}

func buildEnvFileValues(ctx *config.RuntimeConfig) (map[string]string, error) {
	merged := map[string]string{}

	defaultValues, err := config.LoadDefaultEnvValues(ctx.DefaultEnvPath)
	if err != nil {
		return nil, err
	}
	for key, value := range defaultValues {
		merged[key] = value
	}

	envFilePath := ctx.EnvFilePath
	if envFilePath != "" {
		fileValues, readErr := godotenv.Read(envFilePath)
		if readErr != nil {
			return nil, readErr
		}
		for key, value := range fileValues {
			merged[key] = value
		}
	}

	return merged, nil
}
