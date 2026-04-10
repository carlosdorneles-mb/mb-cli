// Package update_test holds integration tests that build the full CLI root (e.g. nested
// mb tools --update-all during mb update).
package update_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cliroot "mb/internal/cli/root"
	"mb/internal/deps"
	"mb/internal/infra/executor"
	"mb/internal/infra/plugins"
	"mb/internal/infra/sqlite"
	"mb/internal/ports"
	"mb/internal/shared/config"
)

func TestUpdateOnlyToolsRunsNestedToolsUpdateAll(t *testing.T) {
	tmp := t.TempDir()
	pluginsDir := filepath.Join(tmp, "plugins")
	toolsDir := filepath.Join(pluginsDir, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	uPath := filepath.Join(toolsDir, "u.sh")
	if err := os.WriteFile(uPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write u.sh: %v", err)
	}

	flagsJSON, err := json.Marshal(map[string]plugins.FlagDef{
		"update-all": {Type: "long", Entrypoint: "u.sh", Description: "update"},
	})
	if err != nil {
		t.Fatalf("marshal flags: %v", err)
	}

	cachePath := filepath.Join(tmp, "mb", "cache.db")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}
	store, err := sqlite.NewStore(cachePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	if err := store.UpsertPlugin(sqlite.Plugin{
		CommandPath: "tools",
		CommandName: "tools",
		Description: "Tools umbrella",
		FlagsJSON:   string(flagsJSON),
		ConfigHash:  "t1",
	}); err != nil {
		t.Fatalf("upsert tools: %v", err)
	}

	cfgDir := filepath.Join(tmp, "mb")
	rt := &deps.RuntimeConfig{Paths: deps.Paths{ConfigDir: cfgDir, PluginsDir: pluginsDir}}
	d := deps.NewDependencies(
		rt,
		config.AppConfig{},
		store,
		plugins.NewScanner(pluginsDir),
		executor.New(),
		nil,
		nil,
	)
	rootCmd := cliroot.NewRootCmd(d, &testOSFS{}, &testGitOps{}, &testShellInstaller{}, &testLayoutValidator{})
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetArgs([]string{"update", "--only-tools"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute update --only-tools: %v", err)
	}
}

type testOSFS struct{}

func (testOSFS) RemoveAll(string) error                          { return nil }
func (testOSFS) MkdirAll(string, os.FileMode) error              { return nil }
func (testOSFS) Stat(name string) (os.FileInfo, error)           { return os.Stat(name) }
func (testOSFS) IsNotExist(err error) bool                       { return os.IsNotExist(err) }
func (testOSFS) ReadDir(name string) ([]os.DirEntry, error)      { return os.ReadDir(name) }
func (testOSFS) Getwd() (string, error)                          { return os.Getwd() }

type testGitOps struct{}

func (testGitOps) ParseGitURL(raw string) (string, string, error) {
	if strings.HasPrefix(raw, "https://") || strings.HasPrefix(raw, "git@") {
		return "repo", raw, nil
	}
	return "", "", fmt.Errorf("not a git URL")
}
func (testGitOps) Clone(context.Context, string, string, ports.GitCloneOpts) error {
	return nil
}
func (testGitOps) LatestTag(context.Context, string) (string, error)        { return "", nil }
func (testGitOps) GetVersion(string) (string, error)                        { return "1.0.0", nil }
func (testGitOps) GetCurrentBranch(string) (string, error)                  { return "main", nil }
func (testGitOps) IsGitRepo(string) bool                                    { return false }
func (testGitOps) FetchTags(context.Context, string) error                  { return nil }
func (testGitOps) ListLocalTags(string) ([]string, error)                   { return nil, nil }
func (testGitOps) NewerTag(string, string) (string, bool)                   { return "", false }
func (testGitOps) CheckoutTag(context.Context, string, string) error        { return nil }
func (testGitOps) FetchAndPull(context.Context, string, string) error       { return nil }

type testShellInstaller struct{}

func (testShellInstaller) EnsureShellHelpers(string) (string, error)        { return "", nil }

type testLayoutValidator struct{}

func (testLayoutValidator) ValidatePluginRoot(string) error                 { return nil }
