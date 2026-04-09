package envs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mapSecretStore map[string]string

func (m mapSecretStore) key(group, k string) string { return group + "\x00" + k }

func (m mapSecretStore) Set(group, k, v string) error {
	m[m.key(group, k)] = v
	return nil
}

func (m mapSecretStore) Get(group, k string) (string, error) {
	v, ok := m[m.key(group, k)]
	if !ok {
		return "", errors.New("missing")
	}
	return v, nil
}

func (m mapSecretStore) Delete(group, k string) error {
	delete(m, m.key(group, k))
	return nil
}

func TestUnset_NotFound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p, []byte("OTHER=x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss := mapSecretStore{}
	removed, err := Unset(ss, nil, Paths{DefaultEnvPath: p, ConfigDir: tmp}, "", "MISSING")
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Fatal("expected removed=false when key is absent")
	}
}

func TestUnset_PlainRemovesFromFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p, []byte("FOO=1\nBAR=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss := mapSecretStore{}
	removed, err := Unset(ss, nil, Paths{DefaultEnvPath: p, ConfigDir: tmp}, "", "FOO")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "FOO") {
		t.Fatalf("FOO should be gone: %q", b)
	}
	if !strings.Contains(string(b), "BAR") {
		t.Fatalf("BAR should remain: %q", b)
	}
}

func TestUnset_SecretKeyListRemovesKeyring(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p+".secrets", []byte("API_KEY\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss := mapSecretStore{}
	if err := ss.Set("default", "API_KEY", "secret-value"); err != nil {
		t.Fatal(err)
	}
	removed, err := Unset(ss, nil, Paths{DefaultEnvPath: p, ConfigDir: tmp}, "", "API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected removed=true for secret-only key")
	}
	if _, err := ss.Get("default", "API_KEY"); err == nil {
		t.Fatal("expected keyring entry removed")
	}
	if _, err := os.Stat(p + ".secrets"); !os.IsNotExist(err) {
		t.Fatal("expected .secrets removed when last key")
	}
}

func TestUnset_GroupRemovesEnvFileWhenLastKeyRemoved(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	groupPath := filepath.Join(tmp, ".env.staging")
	if err := os.WriteFile(groupPath, []byte("FOO=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss := mapSecretStore{}
	removed, err := Unset(ss, nil, Paths{DefaultEnvPath: def, ConfigDir: tmp}, "staging", "FOO")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}
	if _, err := os.Stat(groupPath); !os.IsNotExist(err) {
		t.Fatalf("expected .env.staging removed, stat err=%v", err)
	}
}

func TestUnset_GroupRemovesEnvFileWhenLastSecretRemoved(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	groupPath := filepath.Join(tmp, ".env.prod")
	if err := os.WriteFile(groupPath, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(groupPath+".secrets", []byte("API_KEY\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss := mapSecretStore{}
	if err := ss.Set("prod", "API_KEY", "v"); err != nil {
		t.Fatal(err)
	}
	removed, err := Unset(ss, nil, Paths{DefaultEnvPath: def, ConfigDir: tmp}, "prod", "API_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}
	if _, err := os.Stat(groupPath); !os.IsNotExist(err) {
		t.Fatal("expected .env.prod removed")
	}
}

func TestUnset_DefaultEnvFileNotRemovedWhenEmpty(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p, []byte("ONLY=x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ss := mapSecretStore{}
	removed, err := Unset(ss, nil, Paths{DefaultEnvPath: p, ConfigDir: tmp}, "", "ONLY")
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("env.defaults should still exist: %v", err)
	}
}
