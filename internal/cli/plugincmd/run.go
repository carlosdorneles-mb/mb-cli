package plugincmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"mb/internal/deps"
	domainplugin "mb/internal/domain/plugin"
	"mb/internal/ports"
	"mb/internal/shared/pluginutil"
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
	plugin domainplugin.Plugin,
	d deps.Dependencies,
	exec ports.ScriptExecutor,
	pluginRoot string,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		argsToPass := cmd.Flags().Args()
		merged, err := deps.BuildMergedOSEnviron(d, func(m map[string]string) error {
			return mergeManifestEnvIntoFileValues(
				m,
				plugin.PluginDir,
				plugin.EnvFilesJSON,
				d.Runtime,
			)
		})
		if err != nil {
			return err
		}
		merged = appendPluginInvocationEnv(
			merged,
			cmd,
			plugin,
			os.Args,
			d.Runtime.ConfigDir,
			changedLocalPluginFlags(cmd),
		)
		ctx := cmd.Context()
		if d.Runtime.PluginTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
			defer cancel()
		}
		return exec.Run(ctx, plugin, argsToPass, merged, pluginRoot)
	}
}

func runFlagsOnlyCommand(
	plugin domainplugin.Plugin,
	flagsMap map[string]pluginutil.FlagDef,
	d deps.Dependencies,
	exec ports.ScriptExecutor,
	pluginRoot string,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if f := cmd.Flags().Lookup("readme"); f != nil && f.Changed {
			return runReadmeWithGlow(plugin.ReadmePath)
		}
		_ = parseRootVerbosityFlags(cmd, args)
		var changedFlagNames []string
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if _, ok := flagsMap[f.Name]; ok {
				changedFlagNames = append(changedFlagNames, f.Name)
			}
		})

		var chosenFlagEnv []string
		for _, name := range changedFlagNames {
			chosenFlagEnv = append(chosenFlagEnv, flagsMap[name].Envs...)
		}

		var chosenFlag string
		var chosenEntrypoint string
		if len(changedFlagNames) > 0 {
			chosenFlag = changedFlagNames[0]
			chosenEntrypoint = flagsMap[chosenFlag].Entrypoint
		}

		if chosenFlag == "" || chosenEntrypoint == "" {
			if plugin.ExecPath != "" {
				argsToPass := cmd.Flags().Args()
				merged, err := deps.BuildMergedOSEnvironWithExtraInline(
					d,
					func(m map[string]string) error {
						return mergeManifestEnvIntoFileValues(
							m,
							plugin.PluginDir,
							plugin.EnvFilesJSON,
							d.Runtime,
						)
					},
					chosenFlagEnv,
				)
				if err != nil {
					return err
				}
				flagNames := append([]string(nil), changedFlagNames...)
				sort.Strings(flagNames)
				merged = appendPluginInvocationEnv(
					merged,
					cmd,
					plugin,
					os.Args,
					d.Runtime.ConfigDir,
					flagNames,
				)
				ctx := cmd.Context()
				if d.Runtime.PluginTimeout > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
					defer cancel()
				}
				return exec.Run(ctx, plugin, argsToPass, merged, pluginRoot)
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
		pluginType := pluginutil.PluginTypeFromEntrypoint(chosenEntrypoint)
		syntheticPlugin := domainplugin.Plugin{
			CommandPath:  plugin.CommandPath,
			CommandName:  plugin.CommandName,
			ExecPath:     execPath,
			PluginType:   pluginType,
			ConfigHash:   plugin.ConfigHash,
			PluginDir:    baseDir,
			EnvFilesJSON: plugin.EnvFilesJSON,
		}

		merged, err := deps.BuildMergedOSEnvironWithExtraInline(
			d,
			func(m map[string]string) error {
				return mergeManifestEnvIntoFileValues(
					m,
					syntheticPlugin.PluginDir,
					syntheticPlugin.EnvFilesJSON,
					d.Runtime,
				)
			},
			chosenFlagEnv,
		)
		if err != nil {
			return err
		}
		flagNames := append([]string(nil), changedFlagNames...)
		sort.Strings(flagNames)
		merged = appendPluginInvocationEnv(
			merged,
			cmd,
			plugin,
			os.Args,
			d.Runtime.ConfigDir,
			flagNames,
		)
		ctx := cmd.Context()
		if d.Runtime.PluginTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d.Runtime.PluginTimeout)
			defer cancel()
		}
		return exec.Run(ctx, syntheticPlugin, cmd.Flags().Args(), merged, pluginRoot)
	}
}

func mergeManifestEnvIntoFileValues(
	fileValues map[string]string,
	pluginDir string,
	envFilesJSON string,
	rt *deps.RuntimeConfig,
) error {
	vault := pluginutil.ManifestEnvVaultDefault
	if rt != nil && strings.TrimSpace(rt.EnvVault) != "" {
		vault = strings.TrimSpace(rt.EnvVault)
	}
	extra, err := pluginutil.MergeManifestEnvFiles(pluginDir, envFilesJSON, vault)
	if err != nil {
		return err
	}
	for k, v := range extra {
		fileValues[k] = v
	}
	return nil
}
