package envs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectListRows_includesMbcliProjectAllVaults(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	defaultEnv := filepath.Join(configDir, "env.defaults")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(defaultEnv, []byte("A=from-default\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "mbcli.yaml"))
	t.Setenv("MBCLI_PROJECT_ROOT", "")
	y := "envs:\n  B: from-proj\n  staging:\n    C: only-stg\n"
	if err := os.WriteFile(filepath.Join(tmp, "mbcli.yaml"), []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, err := CollectListRows(
		nil,
		nil,
		Paths{DefaultEnvPath: defaultEnv, ConfigDir: configDir},
		"",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	var foundB, foundC bool
	for _, r := range rows {
		if r.Storage != StorageProject {
			continue
		}
		if r.Key == "B" && r.Vault == "project" {
			foundB = true
		}
		if r.Key == "C" && r.Vault == "project/staging" {
			foundC = true
		}
	}
	if !foundB || !foundC {
		t.Fatalf("missing projeto rows: foundB=%v foundC=%v", foundB, foundC)
	}
}

func TestCollectListRows_vaultStagingMergesMbcliDefaultAndNamed(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	defaultEnv := filepath.Join(configDir, "env.defaults")
	staging := filepath.Join(configDir, ".env.staging")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(defaultEnv, []byte("X=def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staging, []byte("Y=file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "mbcli.yaml"))
	y := "envs:\n  Z: root\n  staging:\n    W: inner\n"
	if err := os.WriteFile(filepath.Join(tmp, "mbcli.yaml"), []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, err := CollectListRows(
		nil,
		nil,
		Paths{DefaultEnvPath: defaultEnv, ConfigDir: configDir},
		"staging",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, r := range rows {
		found[r.Key+"|"+r.Vault+"|"+r.Storage] = true
	}
	for _, want := range []string{
		"Y|staging|" + StorageLocal,
		"Z|project|" + StorageProject,
		"W|project/staging|" + StorageProject,
	} {
		if !found[want] {
			t.Fatalf("missing row %q in %#v", want, rows)
		}
	}
}
