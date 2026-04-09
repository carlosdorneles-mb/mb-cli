package plugincmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

// =============================================================================
// newLeafCommand — flags, readme e shorthands reservados
// =============================================================================

func TestNewLeafCommand_FlagsWithReadmeNoPanicWhenPluginUsesR(t *testing.T) {
	flagsJSON := `{"run":{"type":"long","short":"r","entrypoint":"run.sh","description":"run"}}`
	plugin := sqlite.Plugin{
		CommandPath: "dev/bump",
		CommandName: "bump",
		FlagsJSON:   flagsJSON,
		ReadmePath:  "/tmp/mb-readme-test.md",
	}
	d := deps.NewDependencies(&deps.RuntimeConfig{}, config.AppConfig{}, nil, nil, nil, nil, nil)
	cmd := newLeafCommand("bump", plugin, d, executor.New(), "/tmp", false, nil, nil)
	rf := cmd.Flags().Lookup("readme")
	if rf == nil {
		t.Fatal("readme flag missing")
	}
	if rf.Shorthand != "r" {
		t.Fatalf(
			"readme should keep -r (MB registado antes dos flags do plugin), got shorthand %q",
			rf.Shorthand,
		)
	}
	runF := cmd.Flags().Lookup("run")
	if runF == nil || runF.Shorthand != "" {
		t.Fatalf("run flag should be long-only when -r fica com --readme, got %#v", runF)
	}
}

func TestNewLeafCommand_ReservedRootShorthandDropped(t *testing.T) {
	flagsJSON := `{"watch":{"type":"long","short":"v","entrypoint":"w.sh","description":"w"}}`
	plugin := sqlite.Plugin{
		CommandPath: "p/x",
		CommandName: "x",
		FlagsJSON:   flagsJSON,
	}
	d := deps.NewDependencies(&deps.RuntimeConfig{}, config.AppConfig{}, nil, nil, nil, nil, nil)
	global := map[string]struct{}{"v": {}}
	cmd := newLeafCommand("x", plugin, d, executor.New(), "/tmp", false, nil, global)
	f := cmd.Flags().Lookup("watch")
	if f == nil {
		t.Fatal("watch flag missing")
	}
	if f.Shorthand != "" {
		t.Fatalf("want empty shorthand when manifest requests reserved v, got %q", f.Shorthand)
	}
}

// =============================================================================
// Attach + folha — flags-only e campos Cobra (integração)
// =============================================================================

func TestFlagsOnlyWithShort(t *testing.T) {
	flagsWithShort := map[string]plugins.FlagDef{
		"deploy": {
			Type:        "long",
			Short:       "d",
			Entrypoint:  "deploy.sh",
			Description: "Faz o deploy",
		},
		"rollback": {
			Type:        "long",
			Short:       "r",
			Entrypoint:  "rollback.sh",
			Description: "Reverte o último deploy",
		},
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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := sqlite.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Flags-only with short",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "test",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
		t.Fatalf("upsert tools/hello: %v", err)
	}

	cfgDir := filepath.Join(tmp, "mb")
	pluginsDir := filepath.Join(tmp, "mb", "plugins")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)
	rootCmd := testRootCmdForPluginIntegrationTests(&d)

	var doCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
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

	for _, args := range [][]string{{"--deploy"}, {"-d"}} {
		doCmd.Flags().ParseErrorsAllowlist.UnknownFlags = false
		if err := doCmd.Flags().Parse(args); err != nil {
			t.Errorf("Parse(%v): %v", args, err)
		}
		if !deployFlag.Changed {
			t.Errorf("after Parse(%v), deploy flag not changed", args)
		}
		deployFlag.Changed = false
	}
}

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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := sqlite.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Flags-only long only",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "test",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
		t.Fatalf("upsert tools/hello: %v", err)
	}

	cfgDir := filepath.Join(tmp, "mb")
	pluginsDir := filepath.Join(tmp, "mb", "plugins")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)
	rootCmd := testRootCmdForPluginIntegrationTests(&d)

	var doCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
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

func TestCobraPluginFieldsInjected(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	pluginDir := filepath.Join(pluginsDir, "tools", "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(pluginDir, "run.sh"),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o755,
	); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := sqlite.Plugin{
		CommandPath:     "tools/hello",
		CommandName:     "hello",
		Description:     "Short",
		ExecPath:        filepath.Join(pluginDir, "run.sh"),
		PluginType:      "sh",
		ConfigHash:      "h",
		UseTemplate:     "<name>",
		ArgsCount:       1,
		AliasesJSON:     `["x","run"]`,
		Example:         "mb tools hello do",
		LongDescription: "Long desc",
		Deprecated:      "Use newcmd instead.",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}

	cfgDir := filepath.Join(tmp, "mb")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)
	rootCmd := testRootCmdForPluginIntegrationTests(&d)

	var helloCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "tools" {
			for _, sub := range c.Commands() {
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
	if helloCmd.Example != "mb tools hello do" {
		t.Errorf("Example = %q, want %q", helloCmd.Example, "mb tools hello do")
	}
	if helloCmd.Long != "Long desc" {
		t.Errorf("Long = %q, want %q", helloCmd.Long, "Long desc")
	}
}
