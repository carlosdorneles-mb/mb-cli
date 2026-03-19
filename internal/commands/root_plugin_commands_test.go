package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/config"
	"mb/internal/deps"
	"mb/internal/executor"
	"mb/internal/plugins"
)

// Root command tree with dynamically attached plugin commands (e.g. shell completion suggestions).

// TestRootCmdAttachIncludesPluginTools verifies that after AttachDynamicCommands,
// the root command includes plugin commands (so shell completion will suggest them).
func TestRootCmdAttachIncludesPluginTools(t *testing.T) {
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
		CommandPath: "tools/hello",
		CommandName: "hello",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "test",
	}
	if err := store.UpsertPlugin(plugin); err != nil {
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
	)
	root := NewRootCmd(d)

	var found bool
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"root should have a 'tools' command from plugins (completion will suggest it); AttachDynamicCommands may not be adding plugin commands",
		)
	}
}

func TestRootCmdLocalPluginCommandShortContainsLocal(t *testing.T) {
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

	if err := store.UpsertPlugin(cache.Plugin{
		CommandPath: "tools/hello",
		CommandName: "hello",
		Description: "Hello from tools",
		ExecPath:    "/bin/true",
		PluginType:  "sh",
		ConfigHash:  "test",
	}); err != nil {
		t.Fatalf("upsert plugin: %v", err)
	}
	if err := store.UpsertPluginSource(cache.PluginSource{
		InstallDir: "tools",
		LocalPath:  "/path/to/local",
	}); err != nil {
		t.Fatalf("upsert plugin source: %v", err)
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
	root := NewRootCmd(d)

	var toolsCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			toolsCmd = c
			break
		}
	}
	if toolsCmd == nil {
		t.Fatal("root should have 'tools' command")
	}
	var helloCmd *cobra.Command
	for _, c := range toolsCmd.Commands() {
		if c.Name() == "hello" {
			helloCmd = c
			break
		}
	}
	if helloCmd == nil {
		t.Fatal("tools should have 'hello' command")
	}
	if !strings.Contains(helloCmd.Short, "(local)") {
		t.Errorf("local plugin command Short should contain '(local)', got %q", helloCmd.Short)
	}
}
