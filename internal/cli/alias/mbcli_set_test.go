package alias

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"mb/internal/deps"
	"mb/internal/shared/config"
)

func TestAliasSetMbcliYAML_createsAndUpdates(t *testing.T) {
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
	mbcli := filepath.Join(home, "mbcli.yaml")
	t.Setenv("MBCLI_YAML_PATH", mbcli)

	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      cfgDir,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(cfgDir, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)

	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"px", "--yes", "--mbcli-yaml", "--", "echo", "one"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	m, err := deps.ParseMbcliAliases(mbcli)
	if err != nil {
		t.Fatal(err)
	}
	if m["px"].Command[1] != "one" {
		t.Fatalf("%+v", m["px"])
	}

	cmd2 := newSetCmd(d)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{"px", "--yes", "--mbcli-yaml", "--", "echo", "two"})
	cmd2.SetOut(&out)
	cmd2.SetErr(&out)
	if err := cmd2.Execute(); err != nil {
		t.Fatal(err)
	}
	m2, err := deps.ParseMbcliAliases(mbcli)
	if err != nil {
		t.Fatal(err)
	}
	if m2["px"].Command[1] != "two" {
		t.Fatalf("%+v", m2["px"])
	}
}

func TestAliasUnsetMbcliYAML(t *testing.T) {
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
	mbcli := filepath.Join(home, "mbcli.yaml")
	if err := os.WriteFile(mbcli, []byte(`aliases:
  rm:
    command: [true]
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

	cmd := newUnsetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"rm", "--yes", "--mbcli-yaml"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	m, err := deps.ParseMbcliAliases(mbcli)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Fatalf("left %v", m)
	}
}
