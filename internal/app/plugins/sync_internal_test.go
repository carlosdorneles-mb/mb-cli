package plugins

import (
	"testing"

	"mb/internal/domain/plugin"
)

func TestDiffRemovedKeys(t *testing.T) {
	before := map[string]plugin.Plugin{
		"a": {CommandPath: "a", ConfigHash: "1"},
		"b": {CommandPath: "b", ConfigHash: "2"},
	}
	after := map[string]plugin.Plugin{
		"a": {CommandPath: "a", ConfigHash: "1"},
	}
	got := diffRemovedKeys(before, after)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("diffRemovedKeys = %v, want [b]", got)
	}
}

func TestPluginCommandKey(t *testing.T) {
	if pluginCommandKey(plugin.Plugin{CommandPath: "tools/x", CommandName: "y"}) != "tools/x" {
		t.Fatal("want command_path when set")
	}
	if pluginCommandKey(plugin.Plugin{CommandPath: "", CommandName: "leaf"}) != "leaf" {
		t.Fatal("want command_name when path empty")
	}
}
