package run

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
	"mb/internal/shared/config"
)

func testDependencies(t *testing.T) deps.Dependencies {
	t.Helper()
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(tmp, "plugins"),
		},
	}
	return deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
}

func TestNewRunCmd_RequiresArg(t *testing.T) {
	d := testDependencies(t)
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(NewRunCmd(d))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"run"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing command name")
	}
}

func TestNewRunCmd_RunsTrue(t *testing.T) {
	d := testDependencies(t)
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(NewRunCmd(d))
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"run", "true"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRunCmd_ambiguousAliasRequiresEnvVault(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	aliasYAML := `version: 2
aliases:
  dup:
    - "true"
  st:
    dup:
      - "true"
`
	if err := os.WriteFile(alib.FilePath(tmp), []byte(aliasYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(tmp, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(NewRunCmd(d))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"run", "dup"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for ambiguous alias")
	}
	root2 := &cobra.Command{Use: "mb"}
	root2.AddCommand(NewRunCmd(d))
	root2.SetOut(io.Discard)
	root2.SetErr(io.Discard)
	root2.SetArgs([]string{"--env-vault", "st", "run", "dup"})
	if err := root2.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRunCmd_ResolvesAlias(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	aliasYAML := `version: 1
aliases:
  doit:
    command: ["true"]
`
	if err := os.WriteFile(alib.FilePath(tmp), []byte(aliasYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(tmp, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(NewRunCmd(d))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"run", "doit"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRunCmd_CommandNotFound(t *testing.T) {
	d := testDependencies(t)
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(NewRunCmd(d))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"run", "mb-nonexistent-cmd-xyz-12345"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRunCmd_InjectsCwdDotenv(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	dotEnvPath := filepath.Join(tmp, ".env")
	if err := os.WriteFile(dotEnvPath, []byte("MB_RUN_INJECT=hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &deps.RuntimeConfig{
		Paths: deps.Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(tmp, "plugins"),
		},
	}
	d := deps.NewDependencies(rt, config.AppConfig{}, nil, nil, nil, nil, nil)
	root := &cobra.Command{Use: "mb"}
	root.AddCommand(NewRunCmd(d))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"run", "sh", "-c", `printf '%s' "$MB_RUN_INJECT"`})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origOut := os.Stdout
	os.Stdout = w
	execErr := root.Execute()
	_ = w.Close()
	os.Stdout = origOut
	if execErr != nil {
		t.Fatal(execErr)
	}
	body, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "hello" {
		t.Fatalf("stdout=%q want hello", got)
	}
}
