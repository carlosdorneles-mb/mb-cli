package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/cache"
	"mb/internal/commands/config"
	"mb/internal/executor"
	"mb/internal/plugins"
)

// TestCompletionIncludesPluginCommands verifies that after AttachDynamicCommands,
// the root command includes plugin commands (so shell completion will suggest them).
func TestCompletionIncludesPluginCommands(t *testing.T) {
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

	var found bool
	for _, c := range root.Commands() {
		if c.Name() == "tools" {
			found = true
			break
		}
	}
	if !found {
		t.Error("root should have a 'tools' command from plugins (completion will suggest it); AttachDynamicCommands may not be adding plugin commands")
	}
}

func TestLocalPluginCommandShortContainsLocal(t *testing.T) {
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
