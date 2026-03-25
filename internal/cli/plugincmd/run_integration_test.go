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

func envDumpAsMap(t *testing.T, filePath string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read env dump: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	out := map[string]string{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		out[parts[0]] = parts[1]
	}
	return out
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

func TestEntrypointAndFlagsInjectFlagEnvsOnlyWhenFlagProvided(t *testing.T) {
	flagsMap := map[string]plugins.FlagDef{
		"deploy": {
			Type:       "long",
			Short:      "d",
			Entrypoint: "deploy.sh",
			Envs:       []string{"FLAG_A=from-flag"},
		},
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
	script := "#!/bin/sh\nout=\"$1\"\necho \"FLAG_A=${FLAG_A}\" > \"$out\"\n"
	if err := os.WriteFile(
		filepath.Join(pluginDoDir, "run.sh"),
		[]byte(script),
		0o755,
	); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(pluginDoDir, "deploy.sh"),
		[]byte(script),
		0o755,
	); err != nil {
		t.Fatalf("write deploy.sh: %v", err)
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
		CommandPath: "tools/do",
		CommandName: "do",
		Description: "Default + flags",
		ExecPath:    filepath.Join(pluginDoDir, "run.sh"),
		PluginType:  "sh",
		ConfigHash:  "test",
		FlagsJSON:   string(flagsJSON),
		PluginDir:   pluginDoDir,
	}); err != nil {
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

	withoutFlag := filepath.Join(tmp, "without-flag.txt")
	rootCmd.SetArgs([]string{"tools", "do", withoutFlag})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute without flag: %v", err)
	}
	gotNoFlag := envDumpAsMap(t, withoutFlag)
	if gotNoFlag["FLAG_A"] != "" {
		t.Fatalf("FLAG_A without flag = %q, want empty", gotNoFlag["FLAG_A"])
	}

	withFlag := filepath.Join(tmp, "with-flag.txt")
	rootCmd = root.NewRootCmd(d)
	rootCmd.SetArgs([]string{"tools", "do", "--deploy", withFlag})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute with flag: %v", err)
	}
	gotWithFlag := envDumpAsMap(t, withFlag)
	if gotWithFlag["FLAG_A"] != "from-flag" {
		t.Fatalf("FLAG_A with --deploy = %q, want from-flag", gotWithFlag["FLAG_A"])
	}
}

func TestFlagsOnlyMergeFlagEnvsWithPrecedence(t *testing.T) {
	flagsMap := map[string]plugins.FlagDef{
		"deploy": {
			Type:       "long",
			Entrypoint: "run.sh",
			Envs:       []string{"FLAG_A=from-deploy", "COMMON=from-deploy"},
		},
		"rollback": {
			Type:       "long",
			Entrypoint: "run.sh",
			Envs:       []string{"FLAG_B=from-rollback", "COMMON=from-rollback"},
		},
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
	if err := os.WriteFile(
		filepath.Join(pluginDoDir, ".env.default"),
		[]byte("GLOBAL_ONLY=from-file\nCOMMON=from-file\n"),
		0o644,
	); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	script := "#!/bin/sh\nout=\"$1\"\n{\n  echo \"GLOBAL_ONLY=${GLOBAL_ONLY}\"\n  echo \"FLAG_A=${FLAG_A}\"\n  echo \"FLAG_B=${FLAG_B}\"\n  echo \"COMMON=${COMMON}\"\n} > \"$out\"\n"
	if err := os.WriteFile(
		filepath.Join(pluginDoDir, "run.sh"),
		[]byte(script),
		0o755,
	); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	envFilesJSONBytes, err := json.Marshal(
		[]plugins.EnvFileEntry{{File: ".env.default", Group: "default"}},
	)
	if err != nil {
		t.Fatalf("marshal env_files: %v", err)
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
		CommandPath:  "tools/do",
		CommandName:  "do",
		Description:  "Flags-only env merge",
		FlagsJSON:    string(flagsJSON),
		ConfigHash:   "test",
		PluginDir:    pluginDoDir,
		EnvFilesJSON: string(envFilesJSONBytes),
	}); err != nil {
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

	out := filepath.Join(tmp, "env-out.txt")
	rootCmd.SetArgs(
		[]string{"tools", "do", "--deploy", "--rollback", "--env", "COMMON=from-cli", out},
	)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := envDumpAsMap(t, out)
	if got["GLOBAL_ONLY"] != "from-file" {
		t.Fatalf("GLOBAL_ONLY=%q, want from-file", got["GLOBAL_ONLY"])
	}
	if got["FLAG_A"] != "from-deploy" {
		t.Fatalf("FLAG_A=%q, want from-deploy", got["FLAG_A"])
	}
	if got["FLAG_B"] != "from-rollback" {
		t.Fatalf("FLAG_B=%q, want from-rollback", got["FLAG_B"])
	}
	if got["COMMON"] != "from-cli" {
		t.Fatalf("COMMON=%q, want from-cli", got["COMMON"])
	}
}
