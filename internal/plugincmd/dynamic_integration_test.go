package plugincmd_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands"
	"mb/internal/deps"
	"mb/internal/executor"
	"mb/internal/plugins"
	"mb/internal/shared/config"
)

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
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := cache.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Flags-only with short",
		FlagsJSON:   string(flagsJSON),
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

	cfgDir := filepath.Join(tmp, "mb")
	pluginsDir := filepath.Join(tmp, "mb", "plugins")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	root := commands.NewRootCmd(d)

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

	cfgDir := filepath.Join(tmp, "mb")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	root := commands.NewRootCmd(d)

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

	doCmd.SetContext(context.Background())
	doCmd.Flags().ParseErrorsAllowlist.UnknownFlags = false
	if err := doCmd.Flags().Parse([]string{}); err != nil {
		t.Fatalf("Parse(): %v", err)
	}
	if err := doCmd.RunE(doCmd, []string{}); err != nil {
		t.Errorf("RunE with no flag (default entrypoint): %v", err)
	}

	doCmd.Flags().Parse([]string{"--deploy"})
	if err := doCmd.RunE(doCmd, []string{}); err != nil {
		t.Errorf("RunE with --deploy: %v", err)
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
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := cache.Plugin{
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Flags-only long only",
		FlagsJSON:   string(flagsJSON),
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

	cfgDir := filepath.Join(tmp, "mb")
	pluginsDir := filepath.Join(tmp, "mb", "plugins")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	root := commands.NewRootCmd(d)

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
	)
	root := commands.NewRootCmd(d)

	var helloCmd *cobra.Command
	for _, c := range root.Commands() {
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

func TestEntrypointCommandHelpShowsHelp(t *testing.T) {
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
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		Description: "Short",
		ExecPath:    filepath.Join(pluginDir, "run.sh"),
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
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
	)
	root := commands.NewRootCmd(d)
	var out strings.Builder
	root.SetOut(&out)
	root.SetArgs([]string{"tools", "hello", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected help output, got nothing")
	}
	if !strings.Contains(out.String(), "Usage") && !strings.Contains(out.String(), "Short") {
		t.Errorf("help output should contain Usage or Short, got: %s", out.String())
	}
}

func TestEntrypointCommandGlobalFlagsNotPassedToPlugin(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	pluginDir := filepath.Join(pluginsDir, "tools", "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	argsFile := filepath.Join(tmp, "args.txt")
	script := "#!/bin/sh\nout=\"$1\"\nshift\necho \"$@\" > \"$out\"\n"
	if err := os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte(script), 0o755); err != nil {
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
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    filepath.Join(pluginDir, "run.sh"),
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
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
	)
	root := commands.NewRootCmd(d)
	root.SetArgs([]string{"tools", "hello", "-v", argsFile, "foo"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	raw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read args file: %v", err)
	}
	got := strings.TrimSpace(string(raw))
	if strings.Contains(got, "-v") {
		t.Errorf("plugin should not receive -v, got %q", got)
	}
	if got != "foo" {
		t.Errorf("plugin args = %q, want %q", got, "foo")
	}
}

func TestEntrypointCommandPositionalArgsPassedToPlugin(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	pluginDir := filepath.Join(pluginsDir, "tools", "hello")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	argsFile := filepath.Join(tmp, "args.txt")
	script := "#!/bin/sh\nout=\"$1\"\nshift\necho \"$@\" > \"$out\"\n"
	if err := os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte(script), 0o755); err != nil {
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
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    filepath.Join(pluginDir, "run.sh"),
		PluginType:  "sh",
		ConfigHash:  "h",
	}); err != nil {
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
	)
	root := commands.NewRootCmd(d)
	root.SetArgs([]string{"tools", "hello", argsFile, "foo", "bar"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	raw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read args file: %v", err)
	}
	got := strings.TrimSpace(string(raw))
	if got != "foo bar" {
		t.Errorf("plugin args = %q, want %q", got, "foo bar")
	}
}

// Repro: leaf manifest at tools/ (flags-only) plus tools/bruno with group_id from groups.yaml
// used to panic: group id 'development' is not defined for subcommand 'mb tools bruno'.
func TestLeafToolsWithNestedBrunoHelpGroupNoPanic(t *testing.T) {
	flagsJSON, err := json.Marshal(map[string]plugins.FlagDef{
		"update-all": {Type: "long", Entrypoint: "u.sh", Description: "update"},
	})
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	brunoDir := filepath.Join(pluginsDir, "tools", "bruno")
	if err := os.MkdirAll(brunoDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(brunoDir, "index.sh"),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o755,
	); err != nil {
		t.Fatalf("write index.sh: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	store, err := cache.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPluginHelpGroup(cache.PluginHelpGroup{
		GroupID: "development",
		Title:   "Desenvolvimento",
	}); err != nil {
		t.Fatalf("upsert help group: %v", err)
	}
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools",
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("upsert tools leaf: %v", err)
	}
	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/bruno",
		CommandName: "bruno",
		Description: "Bruno",
		ExecPath:    filepath.Join(brunoDir, "index.sh"),
		PluginType:  "sh",
		ConfigHash:  "b1",
		GroupID:     "development",
		PluginDir:   brunoDir,
	}); err != nil {
		t.Fatalf("upsert tools/bruno: %v", err)
	}

	cfgDir := filepath.Join(tmp, "mb")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	root := commands.NewRootCmd(d)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Execute panicked (Cobra group mismatch): %v", r)
		}
	}()
	root.SetArgs([]string{})
	var out strings.Builder
	root.SetOut(&out)
	root.SetErr(&out)
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}
