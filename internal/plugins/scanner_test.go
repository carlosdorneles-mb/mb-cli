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

func TestScannerCobraFieldsFromManifest(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "tools", "mycmd")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	manifest := []byte(`command: mycmd
description: Short
long_description: "Long desc"
entrypoint: run.sh
use: "<name> [options]"
args: 1
aliases:
  - x
  - run
example: "mb tools mycmd dudu"
deprecated: "Use newcmd instead."
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, _, warnings, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: %v", warnings)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	p := plugins[0]
	if p.UseTemplate != "<name> [options]" || p.ArgsCount != 1 {
		t.Errorf("use=%q args=%d", p.UseTemplate, p.ArgsCount)
	}
	if p.AliasesJSON != `["x","run"]` {
		t.Errorf("aliases_json=%q", p.AliasesJSON)
	}
	if p.Example != "mb tools mycmd dudu" || p.LongDescription != "Long desc" || p.Deprecated != "Use newcmd instead." {
		t.Errorf("example=%q long=%q deprecated=%q", p.Example, p.LongDescription, p.Deprecated)
	}
}

func TestScannerRejectsMissingCommandForEntrypoint(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "tools", "no-cmd")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "run.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write run.sh: %v", err)
	}
	// entrypoint set but command missing -> validation error
	manifest := []byte("description: No command\nentrypoint: run.sh\n")
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, _, warnings, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("expected no plugins when command missing, got %d", len(plugins))
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Message, "command") || !strings.Contains(warnings[0].Message, "obrigatório") {
		t.Errorf("warning should mention command obrigatório, got %q", warnings[0].Message)
	}
}

func TestScannerRejectsMissingCommandForFlags(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "tools", "flags-only")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "deploy.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write deploy.sh: %v", err)
	}
	manifest := []byte(`description: Flags only, no command
flags:
  - name: deploy
    description: Deploy
    entrypoint: deploy.sh
    commands:
      long: deploy
`)
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, _, warnings, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("expected no plugins when command missing (flags-only), got %d", len(plugins))
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Message, "command") || !strings.Contains(warnings[0].Message, "obrigatório") {
		t.Errorf("warning should mention command obrigatório, got %q", warnings[0].Message)
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

func TestScannerRejectsEntrypointPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "tools", "safe")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Script outside plugin dir: tmp/other/run.sh (so path escapes plugins/tools/safe)
	outsideDir := filepath.Join(tmp, "other")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	outsideScript := filepath.Join(outsideDir, "run.sh")
	if err := os.WriteFile(outsideScript, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write outside script: %v", err)
	}
	// Manifest points to script outside plugin dir via .. (safe -> tools -> plugins -> tmp, then other/run.sh)
	manifest := []byte("command: safe\ndescription: Safe\nentrypoint: ../../../other/run.sh\n")
	if err := os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	scanner := NewScanner(filepath.Join(tmp, "plugins"))
	plugins, _, warnings, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("expected no plugins (path traversal), got %d", len(plugins))
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0].Message, "fora do diretório") {
		t.Errorf("warning should mention path outside plugin dir, got %q", warnings[0].Message)
	}
}

func TestScannerEntrypointAndFlags(t *testing.T) {
	tmp := t.TempDir()
	pluginDir := filepath.Join(tmp, "plugins", "tools", "do")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runPath := filepath.Join(pluginDir, "run.sh")
	deployPath := filepath.Join(pluginDir, "deploy.sh")
	for _, p := range []string{runPath, deployPath} {
		if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write script: %v", err)
		}
	}
	manifest := []byte(`command: do
description: Do with default and flags
entrypoint: run.sh
flags:
  - name: deploy
    description: Deploy
    entrypoint: deploy.sh
    commands:
      long: deploy
      short: d
`)
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
	p := plugins[0]
	if p.ExecPath != runPath {
		t.Errorf("ExecPath = %q, want %q", p.ExecPath, runPath)
	}
	if p.FlagsJSON == "" {
		t.Error("FlagsJSON should be set when manifest has both entrypoint and flags")
	}
	if p.PluginType != "sh" {
		t.Errorf("PluginType = %q, want sh", p.PluginType)
	}
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
