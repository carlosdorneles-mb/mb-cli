package plugins

import (
	"path/filepath"
	"testing"

	"mb/internal/cache"
)

func TestFirstPathSegment(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"foo", "foo"},
		{"a/b/c", "a"},
	}
	for _, tt := range tests {
		if got := FirstPathSegment(tt.in); got != tt.want {
			t.Errorf("FirstPathSegment(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPluginDirUnderRoot(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "plugins", "p1")
	if !PluginDirUnderRoot(root, filepath.Join(root, "sub")) {
		t.Error("sub should be under root")
	}
	if PluginDirUnderRoot(root, filepath.Join(root, "..", "p2")) {
		t.Error("sibling should not be under root")
	}
	if PluginDirUnderRoot("", root) {
		t.Error("empty root")
	}
}

func TestSourceForPluginByInstallDir(t *testing.T) {
	sources := []cache.PluginSource{
		{InstallDir: "tools", GitURL: "https://x"},
	}
	p := cache.Plugin{CommandPath: "tools/do", PluginDir: ""}
	got := SourceForPlugin(p, sources, "/plugins")
	if got == nil || got.InstallDir != "tools" {
		t.Fatalf("got %+v", got)
	}
}

func TestSourceForPluginLocalPathLongest(t *testing.T) {
	localA := filepath.Join(string(filepath.Separator), "proj", "plugins")
	localB := filepath.Join(localA, "nested")
	sources := []cache.PluginSource{
		{InstallDir: "short", LocalPath: localA},
		{InstallDir: "long", LocalPath: localB},
	}
	p := cache.Plugin{PluginDir: filepath.Join(localB, "leaf")}
	got := SourceForPlugin(p, sources, "/x")
	if got == nil || got.InstallDir != "long" {
		t.Fatalf("want longest LocalPath match, got %+v", got)
	}
}
