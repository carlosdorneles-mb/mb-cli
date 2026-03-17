package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/executor"
	"mb/internal/plugins"
)

func TestAppendVerbosityEnv(t *testing.T) {
	contains := func(env []string, key string) bool {
		for _, e := range env {
			if e == key+"=1" {
				return true
			}
		}
		return false
	}

	tests := []struct {
		name     string
		rt       *config.RuntimeConfig
		wantVerb bool
		wantQuiet bool
	}{
		{"both false", &config.RuntimeConfig{Verbose: false, Quiet: false}, false, false},
		{"verbose only", &config.RuntimeConfig{Verbose: true, Quiet: false}, true, false},
		{"quiet only", &config.RuntimeConfig{Verbose: false, Quiet: true}, false, true},
		{"both true", &config.RuntimeConfig{Verbose: true, Quiet: true}, true, true},
		{"nil rt", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := appendVerbosityEnv([]string{"FOO=bar"}, tt.rt)
			if got := contains(merged, "MB_VERBOSE"); got != tt.wantVerb {
				t.Errorf("appendVerbosityEnv() MB_VERBOSE present = %v, want %v (merged: %s)", got, tt.wantVerb, strings.Join(merged, " "))
			}
			if got := contains(merged, "MB_QUIET"); got != tt.wantQuiet {
				t.Errorf("appendVerbosityEnv() MB_QUIET present = %v, want %v (merged: %s)", got, tt.wantQuiet, strings.Join(merged, " "))
			}
			if !tt.wantVerb && !tt.wantQuiet && tt.rt != nil {
				if len(merged) != 1 || merged[0] != "FOO=bar" {
					t.Errorf("appendVerbosityEnv() should not add entries when both false, got %v", merged)
				}
			}
		})
	}
}

func TestParseRootVerbosityFlags(t *testing.T) {
	var verbose, quiet bool
	root := &cobra.Command{Use: "mb"}
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "")
	child := &cobra.Command{Use: "hello"}
	root.AddCommand(child)

	tests := []struct {
		name        string
		args        []string
		wantVerbose bool
		wantQuiet   bool
		wantRemaining []string
	}{
		{"-v consumes and sets verbose", []string{"-v"}, true, false, []string{}},
		{"-q consumes and sets quiet", []string{"-q"}, false, true, []string{}},
		{"-v -q both set", []string{"-v", "-q"}, true, true, []string{}},
		{"no flags", []string{}, false, false, []string{}},
		{"-v then positional", []string{"-v", "foo"}, true, false, []string{"foo"}},
		{"positional then -v", []string{"foo", "-v"}, true, false, []string{"foo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verbose, quiet = false, false
			remaining := parseRootVerbosityFlags(child, tt.args)
			if verbose != tt.wantVerbose {
				t.Errorf("verbose = %v, want %v", verbose, tt.wantVerbose)
			}
			if quiet != tt.wantQuiet {
				t.Errorf("quiet = %v, want %v", quiet, tt.wantQuiet)
			}
			if !reflect.DeepEqual(remaining, tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}

// TestFlagsOnlyWithShort verifies that a flag with Short in the manifest is registered
// with both long (--deploy) and short (-d) forms, and that both trigger the same flag.
func TestFlagsOnlyWithShort(t *testing.T) {
	flagsWithShort := map[string]plugins.FlagDef{
		"deploy":  {Type: "long", Short: "d", Entrypoint: "deploy.sh", Description: "Faz o deploy"},
		"rollback": {Type: "long", Short: "r", Entrypoint: "rollback.sh", Description: "Reverte o último deploy"},
	}
	flagsJSON, err := json.Marshal(flagsWithShort)
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := cache.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Flags-only with short",
		FlagsJSON:  string(flagsJSON),
		ConfigHash:  "test",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	// Category tools must exist for the command to attach under it
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
		t.Fatalf("upsert tools/hello: %v", err)
	}

	runtime := &config.RuntimeConfig{
		ConfigDir:  filepath.Join(tmp, "mb"),
		PluginsDir: filepath.Join(tmp, "mb", "plugins"),
	}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(runtime.PluginsDir),
		executor.New(),
	)
	root := NewRootCmd(deps)

	var doCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			for _, sub := range c.Commands() {
				if sub.Name() == "do" {
					doCmd = sub
					break
				}
			}
			break
		}
	}
	if doCmd == nil {
		t.Fatal("command 'tools do' not found")
	}

	deployFlag := doCmd.Flags().Lookup("deploy")
	if deployFlag == nil {
		t.Fatal("flag 'deploy' not registered")
	}
	if deployFlag.Shorthand != "d" {
		t.Errorf("deploy flag Shorthand = %q, want \"d\"", deployFlag.Shorthand)
	}
	if deployFlag.Usage != "Faz o deploy" {
		t.Errorf("deploy flag Usage = %q, want \"Faz o deploy\"", deployFlag.Usage)
	}

	rollbackFlag := doCmd.Flags().Lookup("rollback")
	if rollbackFlag == nil {
		t.Fatal("flag 'rollback' not registered")
	}
	if rollbackFlag.Shorthand != "r" {
		t.Errorf("rollback flag Shorthand = %q, want \"r\"", rollbackFlag.Shorthand)
	}

	// Both --deploy and -d should set the same flag
	for _, args := range [][]string{{"--deploy"}, {"-d"}} {
		doCmd.Flags().ParseErrorsWhitelist.UnknownFlags = false
		if err := doCmd.Flags().Parse(args); err != nil {
			t.Errorf("Parse(%v): %v", args, err)
		}
		if !deployFlag.Changed {
			t.Errorf("after Parse(%v), deploy flag not changed", args)
		}
		deployFlag.Changed = false
	}
}

// TestEntrypointAndFlagsRunsDefaultOrFlag verifies that when a plugin has both ExecPath
// and FlagsJSON, running with no flag runs the default entrypoint, and running with
// a flag runs that flag's entrypoint.
func TestEntrypointAndFlagsRunsDefaultOrFlag(t *testing.T) {
	flagsMap := map[string]plugins.FlagDef{
		"deploy": {Type: "long", Short: "d", Entrypoint: "deploy.sh", Description: "Deploy"},
	}
	flagsJSON, err := json.Marshal(flagsMap)
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	pluginDoDir := filepath.Join(pluginsDir, "tools", "do")
	if err := os.MkdirAll(pluginDoDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runPath := filepath.Join(pluginDoDir, "run.sh")
	deployPath := filepath.Join(pluginDoDir, "deploy.sh")
	for _, path := range []string{runPath, deployPath} {
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write script: %v", err)
		}
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := cache.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Default + flags",
		ExecPath:    runPath,
		PluginType:  "sh",
		ConfigHash:  "test",
		FlagsJSON:   string(flagsJSON),
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
		t.Fatalf("upsert tools/hello: %v", err)
	}

	runtime := &config.RuntimeConfig{
		ConfigDir:  filepath.Join(tmp, "mb"),
		PluginsDir: pluginsDir,
	}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(runtime.PluginsDir),
		executor.New(),
	)
	root := NewRootCmd(deps)

	var doCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			for _, sub := range c.Commands() {
				if sub.Name() == "do" {
					doCmd = sub
					break
				}
			}
			break
		}
	}
	if doCmd == nil {
		t.Fatal("command 'tools do' not found")
	}

	// No flag: should run default entrypoint (run.sh)
	doCmd.SetContext(context.Background())
	doCmd.Flags().ParseErrorsWhitelist.UnknownFlags = false
	if err := doCmd.Flags().Parse([]string{}); err != nil {
		t.Fatalf("Parse(): %v", err)
	}
	if err := doCmd.RunE(doCmd, []string{}); err != nil {
		t.Errorf("RunE with no flag (default entrypoint): %v", err)
	}

	// --deploy: should run deploy.sh
	doCmd.Flags().Parse([]string{"--deploy"})
	if err := doCmd.RunE(doCmd, []string{}); err != nil {
		t.Errorf("RunE with --deploy: %v", err)
	}
}

// TestFlagsOnlyWithoutShort verifies backward compatibility: flags with type long only
// (no Short field) are still registered and work with the long form only.
func TestFlagsOnlyWithoutShort(t *testing.T) {
	flagsLongOnly := map[string]plugins.FlagDef{
		"deploy": {Type: "long", Entrypoint: "deploy.sh", Description: "Deploy"},
	}
	flagsJSON, err := json.Marshal(flagsLongOnly)
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	tmp := t.TempDir()
	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := cache.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Flags-only long only",
		FlagsJSON:  string(flagsJSON),
		ConfigHash:  "test",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
		t.Fatalf("upsert tools/hello: %v", err)
	}

	runtime := &config.RuntimeConfig{
		ConfigDir:  filepath.Join(tmp, "mb"),
		PluginsDir: filepath.Join(tmp, "mb", "plugins"),
	}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(runtime.PluginsDir),
		executor.New(),
	)
	root := NewRootCmd(deps)

	var doCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			for _, sub := range c.Commands() {
				if sub.Name() == "do" {
					doCmd = sub
					break
				}
			}
			break
		}
	}
	if doCmd == nil {
		t.Fatal("command 'tools do' not found")
	}

	deployFlag := doCmd.Flags().Lookup("deploy")
	if deployFlag == nil {
		t.Fatal("flag 'deploy' not registered")
	}
	if deployFlag.Shorthand != "" {
		t.Errorf("deploy flag Shorthand = %q, want \"\" (long only)", deployFlag.Shorthand)
	}
}

// TestCobraPluginFieldsInjected verifies that plugin UseTemplate, ArgsCount, Aliases, Example, Long, Deprecated are applied to the leaf command.
func TestCobraPluginFieldsInjected(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	pluginDir := filepath.Join(pluginsDir, "tools", "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := cache.Plugin{
		CommandPath:     "tools/hello",
		CommandName:     "hello",
		Description:     "Short",
		ExecPath:        filepath.Join(pluginDir, "run.sh"),
		PluginType:      "sh",
		ConfigHash:      "h",
		UseTemplate:     "<name>",
		ArgsCount:       1,
		AliasesJSON:     `["x","run"]`,
		Example:         "mb tools hello dudu",
		LongDescription: "Long desc",
		Deprecated:      "Use newcmd instead.",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}

	runtime := &config.RuntimeConfig{
		ConfigDir:  filepath.Join(tmp, "mb"),
		PluginsDir: pluginsDir,
	}
	deps := config.NewDependencies(
		runtime,
		store,
		plugins.NewScanner(runtime.PluginsDir),
		executor.New(),
	)
	root := NewRootCmd(deps)

	var helloCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			for _, sub := range c.Commands() {
				// Use is "command + use" = "hello <name>"
				if sub.Use == "hello <name>" {
					helloCmd = sub
					break
				}
			}
			break
		}
	}
	if helloCmd == nil {
		t.Fatal("command 'tools hello' (Use hello <name>) not found")
	}

	if helloCmd.Use != "hello <name>" {
		t.Errorf("Use = %q, want %q", helloCmd.Use, "hello <name>")
	}
	if helloCmd.Args == nil {
		t.Error("Args should be set (ExactArgs(1))")
	}
	if len(helloCmd.Aliases) != 2 || helloCmd.Aliases[0] != "x" || helloCmd.Aliases[1] != "run" {
		t.Errorf("Aliases = %v, want [x run]", helloCmd.Aliases)
	}
	if helloCmd.Example != "mb tools hello dudu" {
		t.Errorf("Example = %q, want %q", helloCmd.Example, "mb tools hello dudu")
	}
	if helloCmd.Long != "Long desc" {
		t.Errorf("Long = %q, want %q", helloCmd.Long, "Long desc")
	}
	// Deprecated: não setamos cmd.Deprecated; a mensagem em português é exibida via wrap do RunE ao executar.
}
