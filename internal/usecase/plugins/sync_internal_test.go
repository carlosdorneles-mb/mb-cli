package plugins

import (
	"testing"

	"mb/internal/domain/plugin"
	"mb/internal/shared/syncdiff"
)

func TestDiffRemovedKeys(t *testing.T) {
	before := map[string]plugin.Plugin{
		"a": {CommandPath: "a", ConfigHash: "1"},
		"b": {CommandPath: "b", ConfigHash: "2"},
	}
	after := map[string]plugin.Plugin{
		"a": {CommandPath: "a", ConfigHash: "1"},
	}
	got := syncdiff.DiffRemovedKeys(before, after)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("DiffRemovedKeys = %v, want [b]", got)
	}
}

func TestPluginCommandKey(t *testing.T) {
	if syncdiff.PluginCommandKey(
		plugin.Plugin{CommandPath: "tools/x", CommandName: "y"},
	) != "tools/x" {
		t.Fatal("want command_path when set")
	}
	if syncdiff.PluginCommandKey(plugin.Plugin{CommandPath: "", CommandName: "leaf"}) != "leaf" {
		t.Fatal("want command_name when path empty")
	}
}
