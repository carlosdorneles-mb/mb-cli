package plugincmd_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cli/root"
	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := sqlite.Plugin{
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
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
	)
	rootCmd := root.NewRootCmd(d)

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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()
	if err := store.UpsertPlugin(sqlite.Plugin{
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
	rootCmd := root.NewRootCmd(d)
	var out strings.Builder
	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{"tools", "hello", "--help"})
	if err := rootCmd.Execute(); err != nil {
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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()
	if err := store.UpsertPlugin(sqlite.Plugin{
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
	rootCmd := root.NewRootCmd(d)
	rootCmd.SetArgs([]string{"tools", "hello", "-v", argsFile, "foo"})
	if err := rootCmd.Execute(); err != nil {
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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()
	if err := store.UpsertPlugin(sqlite.Plugin{
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
	rootCmd := root.NewRootCmd(d)
	rootCmd.SetArgs([]string{"tools", "hello", argsFile, "foo", "bar"})
	if err := rootCmd.Execute(); err != nil {
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
