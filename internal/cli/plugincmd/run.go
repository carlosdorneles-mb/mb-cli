package plugincmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/safepath"
)

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
	plugin sqlite.Plugin,
	d deps.Dependencies,
	pluginRoot string,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		argsToPass := cmd.Flags().Args()
		merged, err := deps.BuildMergedOSEnviron(d, func(m map[string]string) error {
			return mergeManifestEnvIntoFileValues(m, plugin, d.Runtime)
		})
		if err != nil {
			return err
		}
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
	plugin sqlite.Plugin,
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
				merged, err := deps.BuildMergedOSEnviron(d, func(m map[string]string) error {
					return mergeManifestEnvIntoFileValues(m, plugin, d.Runtime)
				})
				if err != nil {
					return err
				}
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
		syntheticPlugin := sqlite.Plugin{
			CommandPath:  plugin.CommandPath,
			CommandName:  plugin.CommandName,
			ExecPath:     execPath,
			PluginType:   pluginType,
			ConfigHash:   plugin.ConfigHash,
			PluginDir:    baseDir,
			EnvFilesJSON: plugin.EnvFilesJSON,
		}

		merged, err := deps.BuildMergedOSEnviron(d, func(m map[string]string) error {
			return mergeManifestEnvIntoFileValues(m, syntheticPlugin, d.Runtime)
		})
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		if d.Runtime.PluginTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
			defer cancel()
		}
		return d.Executor.Run(ctx, syntheticPlugin, cmd.Flags().Args(), merged, pluginRoot)
	}
}

func mergeManifestEnvIntoFileValues(
	fileValues map[string]string,
	plugin sqlite.Plugin,
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
