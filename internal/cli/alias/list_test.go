package alias

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"mb/internal/deps"
	"mb/internal/shared/config"
)

func TestAliasList_JSON(t *testing.T) {
	home := t.TempDir()
	if err := os.Chdir(home); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(home, "mbcfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	aliasesPath := filepath.Join(cfgDir, "aliases.yaml")
	raw := []byte(`version: 1
aliases:
  zed:
    command: ["echo", "z"]
  alpha:
    command: ["docker", "compose", "up"]
    env_vault: staging
`)
	if err := os.WriteFile(aliasesPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      cfgDir,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(cfgDir, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)

	cmd := newListCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--json"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
	// Sorted by name: alpha, zed
	if g, _ := got[0]["name"].(string); g != "alpha" {
		t.Fatalf("first name want alpha got %v", got[0]["name"])
	}
	if g, _ := got[0]["source"].(string); g != "config" {
		t.Fatalf("alpha source: %v", got[0]["source"])
	}
	if g, _ := got[0]["envVault"].(string); g != "staging" {
		t.Fatalf("alpha envVault: %v", got[0]["envVault"])
	}
	cmdArr, ok := got[0]["command"].([]any)
	if !ok || len(cmdArr) != 3 {
		t.Fatalf("alpha command: %#v", got[0]["command"])
	}
	if g, _ := got[1]["name"].(string); g != "zed" {
		t.Fatalf("second name want zed got %v", got[1]["name"])
	}
	if g, _ := got[1]["source"].(string); g != "config" {
		t.Fatalf("zed source: %v", got[1]["source"])
	}
}

func TestAliasList_JSON_projectOverridesGlobal(t *testing.T) {
	home := t.TempDir()
	if err := os.Chdir(home); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(home, "mbcfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	aliasesPath := filepath.Join(cfgDir, "aliases.yaml")
	if err := os.WriteFile(aliasesPath, []byte(`version: 1
aliases:
  same:
    command: ["echo", "global"]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	mbcli := filepath.Join(home, "mbcli.yaml")
	if err := os.WriteFile(mbcli, []byte(`aliases:
  same:
    command: ["echo", "project"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MBCLI_YAML_PATH", mbcli)

	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      cfgDir,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(cfgDir, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
	cmd := newListCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--json"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 merged row, got %d", len(got))
	}
	if got[0]["name"] != "same" || got[0]["source"] != "project" {
		t.Fatalf("%v", got[0])
	}
	cmdArr, _ := got[0]["command"].([]any)
	if len(cmdArr) != 2 || cmdArr[1] != "project" {
		t.Fatalf("command %#v", got[0]["command"])
	}
}

func TestAliasList_JSON_empty(t *testing.T) {
	home := t.TempDir()
	if err := os.Chdir(home); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(home, "mbcfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	aliasesPath := filepath.Join(cfgDir, "aliases.yaml")
	aliasesYAML := []byte("version: 1\naliases: {}\n")
	if err := os.WriteFile(aliasesPath, aliasesYAML, 0o600); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      cfgDir,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(cfgDir, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
	cmd := newListCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--json"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var arr []any
	if err := json.Unmarshal(out.Bytes(), &arr); err != nil || len(arr) != 0 {
		t.Fatalf("want empty json array, got %q err=%v", out.String(), err)
	}
}
