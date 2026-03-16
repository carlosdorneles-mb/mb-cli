package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScannerFindsManifestPlugins(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "infra", "ci", "deploy")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	scriptPath := filepath.Join(pluginDir, "run.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	manifest := []byte("command: deploy\ndescription: Deploy step\nentrypoint: run.sh\nreadme: README.md\n")
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, _, warnings, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d: %v", len(warnings), warnings)
	}

	if len(plugins) != 1 {
		t.Fatalf("expected one plugin, got %d", len(plugins))
	}
	if plugins[0].CommandPath != "infra/ci/deploy" || plugins[0].CommandName != "deploy" || plugins[0].PluginType != "sh" {
		t.Fatalf("unexpected plugin payload: %#v", plugins[0])
	}
}

func TestScannerSkipsInvalidManifest(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "tools", "broken")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// entrypoint set but file does not exist -> validation error
	manifest := []byte("command: broken\ndescription: Broken plugin\nentrypoint: run.sh\n")
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, categories, warnings, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("expected no plugins (invalid manifest), got %d", len(plugins))
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if warnings[0].Path != filepath.Join(pluginDir, "manifest.yaml") {
		t.Errorf("warning path: got %s", warnings[0].Path)
	}
	if warnings[0].Message == "" {
		t.Errorf("warning message empty")
	}
	if !strings.Contains(warnings[0].Message, "entrypoint") {
		t.Errorf("warning message should mention entrypoint, got %q", warnings[0].Message)
	}
	_ = categories
}

func TestScanDir(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "mydir")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	scriptPath := filepath.Join(pluginDir, "run.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), []byte("command: hello\ndescription: Hi\nentrypoint: run.sh\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	scanner := NewScanner("")
	plugins, categories, warnings, err := scanner.ScanDir(pluginDir, "myinstall")
	if err != nil {
		t.Fatalf("ScanDir: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].CommandPath != "myinstall" {
		t.Errorf("CommandPath want myinstall, got %q", plugins[0].CommandPath)
	}
	if plugins[0].ExecPath != scriptPath {
		t.Errorf("ExecPath want absolute path to run.sh, got %q", plugins[0].ExecPath)
	}
	if !filepath.IsAbs(plugins[0].ExecPath) {
		t.Errorf("ExecPath should be absolute")
	}
	_ = categories
}
