package plugins

import (
	"os"
	"path/filepath"
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

	manifest := []byte("command: deploy\ndescription: Deploy step\ntype: sh\nentrypoint: run.sh\nreadme: README.md\n")
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, _, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(plugins) != 1 {
		t.Fatalf("expected one plugin, got %d", len(plugins))
	}
	if plugins[0].CommandPath != "infra/ci/deploy" || plugins[0].CommandName != "deploy" || plugins[0].PluginType != "sh" {
		t.Fatalf("unexpected plugin payload: %#v", plugins[0])
	}
}
