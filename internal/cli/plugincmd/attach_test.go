package plugincmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cli/completion"
	"mb/internal/cli/plugincmd"
	"mb/internal/cli/root"
	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/shared/config"
)

// =============================================================================
// plugincmd.Attach — help groups e folhas aninhadas
// =============================================================================

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
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPluginHelpGroup(sqlite.PluginHelpGroup{
		GroupID: "development",
		Title:   "Desenvolvimento",
	}); err != nil {
		t.Fatalf("upsert help group: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools",
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("upsert tools leaf: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
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
		nil,
		nil,
	)
	r := root.NewRootCmd(d)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Execute panicked (Cobra group mismatch): %v", r)
		}
	}()
	r.SetArgs([]string{})
	var out strings.Builder
	r.SetOut(&out)
	r.SetErr(&out)
	if err := r.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

// =============================================================================
// plugincmd.Attach — aliases em comandos de categoria (categories.aliases_json)
// =============================================================================

// TestAttachAppliesCategoryAliasesFromCache ensures nested category rows with aliases_json
// produce Cobra Aliases on the intermediate command (e.g. mb ai sk → ai skills).
func TestAttachAppliesCategoryAliasesFromCache(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	addDir := filepath.Join(pluginsDir, "ai", "skills", "add")
	if err := os.MkdirAll(addDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runPath := filepath.Join(addDir, "run.sh")
	if err := os.WriteFile(runPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertCategory(sqlite.Category{
		Path:        "ai",
		Description: "AI categoria",
	}); err != nil {
		t.Fatalf("upsert category ai: %v", err)
	}
	if err := store.UpsertCategory(sqlite.Category{
		Path:        "ai/skills",
		Description: "Skills",
		AliasesJSON: `["sk"]`,
	}); err != nil {
		t.Fatalf("upsert category ai/skills: %v", err)
	}
	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "ai/skills/add",
		CommandName: "add",
		Description: "Add skill",
		ExecPath:    runPath,
		PluginType:  "sh",
		ConfigHash:  "h1",
		PluginDir:   addDir,
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
		nil,
		nil,
	)
	rootCmd := root.NewRootCmd(d)

	var aiCmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "ai" {
			aiCmd = c
			break
		}
	}
	if aiCmd == nil {
		t.Fatal("command 'ai' not found under root")
	}

	var skillsCmd *cobra.Command
	for _, c := range aiCmd.Commands() {
		if c.Name() == "skills" {
			skillsCmd = c
			break
		}
	}
	if skillsCmd == nil {
		t.Fatal("nested command 'skills' not found under ai")
	}

	if len(skillsCmd.Aliases) != 1 || skillsCmd.Aliases[0] != "sk" {
		t.Fatalf("skills command Aliases = %v, want [sk]", skillsCmd.Aliases)
	}
}

// =============================================================================
// plugincmd.Reattach — cache vazio remove comandos de plugin e completion
// =============================================================================

func TestReattachClearsPluginCommandsWhenCacheEmpty(t *testing.T) {
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

	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "test",
	}); err != nil {
		t.Fatalf("upsert plugin: %v", err)
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
	r := root.NewRootCmd(d)

	if !rootHasPluginCommand(r, "tools") {
		t.Fatal("expected root to have plugin top-level 'tools' before cache clear")
	}

	if err := store.DeleteAllPlugins(); err != nil {
		t.Fatalf("delete all plugins: %v", err)
	}

	plugincmd.Reattach(r, d)

	if rootHasPluginCommand(r, "tools") {
		t.Fatal("expected plugin 'tools' to be removed from root after Reattach with empty cache")
	}

	var buf bytes.Buffer
	if err := completion.WriteCompletionScript(r, completion.ShellBash, true, &buf); err != nil {
		t.Fatalf("WriteCompletionScript: %v", err)
	}
	script := buf.String()
	if strings.Contains(script, "tools") && strings.Contains(script, "hello") {
		t.Errorf(
			"completion script should not reference removed plugin tools/hello; got substring match in %d-byte script",
			len(script),
		)
	}
}

func rootHasPluginCommand(rootCmd *cobra.Command, name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name && c.GroupID == "plugin_commands" {
			return true
		}
	}
	return false
}
