package alias

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
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
	if m[alib.StoreKey("", "px")].Command[1] != "one" {
		t.Fatalf("%+v", m[alib.StoreKey("", "px")])
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
	if m2[alib.StoreKey("", "px")].Command[1] != "two" {
		t.Fatalf("%+v", m2[alib.StoreKey("", "px")])
	}
}

func TestAliasSetMbcliYAML_sameCommandNewVault_withYes(t *testing.T) {
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
  py:
    - echo
    - hi
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

	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"py", "--yes", "--mbcli-yaml", "--vault", "staging", "--", "echo", "hi"})
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
	if r := m[alib.StoreKey("", "py")]; r.EnvVault != "" || len(r.Command) != 2 ||
		r.Command[1] != "hi" {
		t.Fatalf("root py=%+v", r)
	}
	got := m[alib.StoreKey("staging", "py")]
	if got.EnvVault != "staging" || len(got.Command) != 2 || got.Command[1] != "hi" {
		t.Fatalf("%+v", got)
	}
	if len(m) != 2 {
		t.Fatalf("want 2 alias slots, got %d", len(m))
	}
}

func TestAliasSetMbcliYAML_vaultProjectSlashStaging(t *testing.T) {
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
  q:
    command: [echo, root]
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

	cmd := newSetCmd(d)
	cmd.SetContext(context.Background())
	cmd.SetArgs(
		[]string{"q", "--yes", "--mbcli-yaml", "--vault", "project/staging", "--", "echo", "leaf"},
	)
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
	if r := m[alib.StoreKey("", "q")]; r.EnvVault != "" || r.Command[1] != "root" {
		t.Fatalf("root q=%+v", r)
	}
	got := m[alib.StoreKey("staging", "q")]
	if got.EnvVault != "staging" || len(got.Command) != 2 || got.Command[1] != "leaf" {
		t.Fatalf("%+v", got)
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
