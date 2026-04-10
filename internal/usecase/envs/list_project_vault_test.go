package envs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectListRows_vaultProjectMbcliOnly(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	defaultEnv := filepath.Join(configDir, "env.defaults")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(defaultEnv, []byte("ONLY=cfg\n"), 0o644); err != nil {
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
	y := "envs:\n  B: yaml\n  staging:\n    C: stg\n"
	if err := os.WriteFile(filepath.Join(tmp, "mbcli.yaml"), []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, err := CollectListRows(
		nil,
		nil,
		Paths{DefaultEnvPath: defaultEnv, ConfigDir: configDir},
		"project",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Key != "B" || rows[0].Vault != "project" {
		t.Fatalf("%#v", rows)
	}
}

func TestCollectListRows_vaultProjectSlashStagingMbcliOnly(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	defaultEnv := filepath.Join(configDir, "env.defaults")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(defaultEnv, []byte("ONLY=cfg\n"), 0o644); err != nil {
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
	y := "envs:\n  B: yaml\n  staging:\n    C: stg\n"
	if err := os.WriteFile(filepath.Join(tmp, "mbcli.yaml"), []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, err := CollectListRows(
		nil,
		nil,
		Paths{DefaultEnvPath: defaultEnv, ConfigDir: configDir},
		"project/staging",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, r := range rows {
		found[r.Key+"@"+r.Vault] = true
	}
	if len(rows) != 1 || !found["C@project/staging"] || found["B@project"] {
		t.Fatalf("want only nested staging keys, got %#v", rows)
	}
}
