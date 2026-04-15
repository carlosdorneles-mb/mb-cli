package alias

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/shared/config"
)

func TestSkipAliasProfileForHelp(t *testing.T) {
	if !skipAliasProfileForHelp([]string{"alias", "--help"}) {
		t.Fatal("expected skip for alias --help")
	}
	if !skipAliasProfileForHelp([]string{"alias", "set", "--help"}) {
		t.Fatal("expected skip for alias set --help")
	}
	if skipAliasProfileForHelp([]string{"alias", "list"}) {
		t.Fatal("expected no skip for alias list")
	}
	if skipAliasProfileForHelp([]string{"-v", "alias", "list"}) {
		t.Fatal("expected no skip")
	}
}

func TestProfileHasExpectedAliasBlock(t *testing.T) {
	line := ". '/tmp/x/shell/aliases.bash'"
	block := appendMarkers(line)
	if !ProfileHasExpectedAliasBlock(block, line) {
		t.Fatal("expected match for fresh block")
	}
	if ProfileHasExpectedAliasBlock("no markers here", line) {
		t.Fatal("expected no match")
	}
	wrong := appendMarkers(". '/old/path'")
	if ProfileHasExpectedAliasBlock(wrong, line) {
		t.Fatal("expected no match for wrong inner line")
	}
}

func TestEnsureShellProfileForAliases_idempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/bin/bash")

	cfgDir := filepath.Join(home, "mbcfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	def := filepath.Join(cfgDir, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
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

	rc := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(rc, []byte("# empty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := &cobra.Command{}
	c.SetContext(context.Background())
	c.SetOut(io.Discard)
	c.SetErr(io.Discard)
	opts := EnsureProfileOptions{}

	if err := ensureShellProfileForAliases(d, c, opts); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}
	wantLine, err := profileSourceLine(cfgDir, "bash")
	if err != nil {
		t.Fatal(err)
	}
	if !ProfileHasExpectedAliasBlock(string(first), wantLine) {
		t.Fatalf("profile missing expected block:\n%s", first)
	}

	if err := ensureShellProfileForAliases(d, c, opts); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("second run changed profile:\nwas:\n%s\nnow:\n%s", first, second)
	}
}
