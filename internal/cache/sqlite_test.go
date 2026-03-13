package cache

import (
	"path/filepath"
	"testing"
)

func TestStoreUpsertAndList(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewStore(filepath.Join(tmp, "cache.db"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	plugin := Plugin{
		CommandPath: "infra/ci/deploy",
		CommandName: "deploy",
		ExecPath:    "/tmp/deploy.sh",
		PluginType:  "sh",
		ConfigHash:  "hash1",
		ReadmePath:  "/tmp/README.md",
	}

	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	plugin.ConfigHash = "hash2"
	if err := store.UpsertPlugin(plugin); err != nil {
		t.Fatalf("upsert second: %v", err)
	}

	all, err := store.ListPlugins()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(all))
	}
	if all[0].ConfigHash != "hash2" {
		t.Fatalf("expected updated hash")
	}

	byPrefix, err := store.ListByPathPrefix("infra")
	if err != nil {
		t.Fatalf("list by prefix: %v", err)
	}
	if len(byPrefix) != 1 {
		t.Fatalf("expected 1 plugin with prefix infra, got %d", len(byPrefix))
	}
}
