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
	"mb/internal/env"
	"mb/internal/plugins"
	"mb/internal/system"
	"mb/internal/ui"
)

func AttachDynamicCommands(root *cobra.Command, deps Dependencies) {
	pluginList, err := deps.Store.ListPlugins()
	if err != nil {
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderError(err.Error()))
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderInfo("run `mb self sync` to rebuild plugin cache"))
		return
	}
	if len(pluginList) == 0 {
		fmt.Fprintln(root.ErrOrStderr(), ui.RenderInfo("no plugins in cache. run `mb self sync` first"))
		return
	}

	categoryList, _ := deps.Store.ListCategories()
	categoriesByPath := make(map[string]cache.Category)
	for _, c := range categoryList {
		categoriesByPath[c.Path] = c
	}

	sort.Slice(pluginList, func(i, j int) bool {
		return pluginList[i].CommandPath < pluginList[j].CommandPath
	})

	// Build tree from command_path: "infra/ci/deploy" -> root -> infra -> ci -> deploy
	type node struct {
		cmd     *cobra.Command
		plugins []cache.Plugin // only at leaf
	}
	byPath := map[string]*node{}

	for _, plugin := range pluginList {
		segments := strings.Split(plugin.CommandPath, "/")
		if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
			// root-level command
			segments = []string{plugin.CommandName}
		}

		var parent *cobra.Command = root
		for i := 0; i < len(segments)-1; i++ {
			seg := segments[i]
			pathSoFar := strings.Join(segments[:i+1], "/")
			if byPath[pathSoFar] == nil {
				cat := categoriesByPath[pathSoFar]
				short := seg + " commands"
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
					categoryCmd.Flags().Bool("readme", false, readmeFlagDesc)
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
			continue // same path already added (should not happen with unique command_path)
		}
		leafCmd := newLeafCommand(plugin.CommandName, plugin, deps)
		if parent == root {
			leafCmd.GroupID = "plugin_commands"
		}
		parent.AddCommand(leafCmd)
		byPath[pathSoFar] = &node{cmd: leafCmd, plugins: []cache.Plugin{plugin}}
	}
}

func pluginDescription(desc, fallback, suffix string) string {
	if desc != "" {
		return desc
	}
	return fallback + " " + suffix
}

func newLeafCommand(use string, plugin cache.Plugin, deps Dependencies) *cobra.Command {
	short := plugin.Description
	if short == "" {
		short = "Run " + plugin.CommandPath
	}

	// Entrypoint plugin: pass-through all args; optional --readme when ReadmePath is set
	if plugin.FlagsJSON == "" {
		cmd := &cobra.Command{
			Use:   use,
			Short: short,
			RunE:  runEntrypointCommand(plugin, deps),
		}
		if plugin.ReadmePath != "" {
			cmd.Flags().Bool("readme", false, readmeFlagDesc)
		} else {
			cmd.DisableFlagParsing = true
		}
		setHelpFang(cmd)
		return cmd
	}

	// Flags-only plugin: register flags from FlagsJSON; no flag -> help; flag -> run entrypoint
	var flagsMap map[string]plugins.FlagDef
	if err := json.Unmarshal([]byte(plugin.FlagsJSON), &flagsMap); err != nil {
		cmd := &cobra.Command{Use: use, Short: short + " (invalid flags config)"}
		return cmd
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  runFlagsOnlyCommand(plugin, flagsMap, deps),
	}

	for name, def := range flagsMap {
		switch def.Type {
		case "long":
			cmd.Flags().Bool(name, false, "")
		case "short":
			if len(name) == 1 {
				cmd.Flags().BoolP(name, name, false, "")
			} else {
				cmd.Flags().Bool(name, false, "")
			}
		}
	}
	if plugin.ReadmePath != "" {
		cmd.Flags().Bool("readme", false, readmeFlagDesc)
	}
	setHelpFang(cmd)
	return cmd
}

func runEntrypointCommand(plugin cache.Plugin, deps Dependencies) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
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
		return deps.Executor.Run(cmd.Context(), plugin, argsToPass, merged)
	}
}

func runFlagsOnlyCommand(plugin cache.Plugin, flagsMap map[string]plugins.FlagDef, deps Dependencies) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		// Check which mapped flag was set
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

		baseDir := filepath.Join(deps.Runtime.PluginsDir, plugin.CommandPath)
		execPath := filepath.Join(baseDir, chosenEntrypoint)
		syntheticPlugin := cache.Plugin{
			CommandPath: plugin.CommandPath,
			CommandName: plugin.CommandName,
			ExecPath:    execPath,
			PluginType:  "sh",
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
		return deps.Executor.Run(cmd.Context(), syntheticPlugin, args, merged)
	}
}

const readmeFlagDesc = "Render README"

func setHelpFang(c *cobra.Command) {
	c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if root := cmd.Root(); root != nil {
			root.HelpFunc()(cmd, args)
		}
	})
}

func runReadmeWithGlow(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return system.RenderMarkdown(context.Background(), path)
}

func buildEnvFileValues(ctx *RuntimeConfig) (map[string]string, error) {
	merged := map[string]string{}

	defaultValues, err := godotenv.Read(ctx.DefaultEnvPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
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
