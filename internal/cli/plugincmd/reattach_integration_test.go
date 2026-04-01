package plugincmd_test

import (
	"bytes"
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
