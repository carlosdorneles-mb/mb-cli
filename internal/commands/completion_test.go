package commands

import (
	"os"
	"path/filepath"
	"testing"

	"mb/internal/cache"
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

	runtime := &RuntimeConfig{
		ConfigDir:  filepath.Join(tmp, "mb"),
		PluginsDir: filepath.Join(tmp, "mb", "plugins"),
	}
	deps := NewDependencies(
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
