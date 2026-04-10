package deps

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type stringSecretStore map[string]string

func (m stringSecretStore) key(group, k string) string { return group + "\x00" + k }

func (m stringSecretStore) Set(group, k, v string) error {
	m[m.key(group, k)] = v
	return nil
}

func (m stringSecretStore) Get(group, k string) (string, error) {
	v, ok := m[m.key(group, k)]
	if !ok {
		return "", errors.New("missing")
	}
	return v, nil
}

func (m stringSecretStore) Delete(group, k string) error {
	delete(m, m.key(group, k))
	return nil
}

type stubOnePassword struct {
	read func(ref string) (string, error)
}

func (s stubOnePassword) EnsureAvailable() error { return nil }

func (s stubOnePassword) PutSecret(_, _, _ string) (string, error) {
	return "", errors.New("not used")
}

func (s stubOnePassword) RemoveSecretField(_, _ string) error { return nil }

func (s stubOnePassword) ReadOPReference(ref string) (string, error) {
	if s.read != nil {
		return s.read(ref)
	}
	return "", errors.New("no read")
}

func TestBuildEnvFileValues_DefaultOnly(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(p, []byte("A=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: p,
		},
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "1" {
		t.Errorf("A=%q", m["A"])
	}
}

func TestBuildEnvFileValues_GroupOverlay(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("A=1\nX=0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	grp := filepath.Join(tmp, ".env.staging")
	if err := os.WriteFile(grp, []byte("A=3\nB=2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
		EnvVault: "staging",
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "3" || m["B"] != "2" || m["X"] != "0" {
		t.Fatalf("merged=%v", m)
	}
}

func TestBuildEnvFileValues_CwdDotEnvBeforeEnvFile(t *testing.T) {
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
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("A=def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dotEnv := filepath.Join(tmp, ".env")
	if err := os.WriteFile(dotEnv, []byte("A=cwd\nB=cwd\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(tmp, "extra.env")
	if err := os.WriteFile(extra, []byte("A=file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
		EnvFilePath: extra,
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "file" {
		t.Errorf("A=%q, want file (--env-file must overlay ./.env)", m["A"])
	}
	if m["B"] != "cwd" {
		t.Errorf("B=%q, want cwd", m["B"])
	}
}

func TestBuildEnvFileValues_MbcliYAMLEnvsOverDefaults(t *testing.T) {
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
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "mbcli.yaml"))
	t.Setenv("MBCLI_PROJECT_ROOT", "")

	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("A=def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(tmp, "mbcli.yaml")
	if err := os.WriteFile(yamlPath, []byte("envs:\n  A: from-yaml\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "from-yaml" {
		t.Fatalf("A=%q, want from-yaml (mbcli.yaml entre default e .env)", m["A"])
	}
}

func TestBuildEnvFileValues_MbcliYAMLNestedVaultWithEnvVault(t *testing.T) {
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
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "mbcli.yaml"))
	t.Setenv("MBCLI_PROJECT_ROOT", "")

	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("FOO=def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  FOO: root\n  staging:\n    FOO: stg\n"
	if err := os.WriteFile(yamlPath, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
		EnvVault: "staging",
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "stg" {
		t.Fatalf("FOO=%q want stg (mbcli staging overlay com --env-vault)", m["FOO"])
	}
}

func TestBuildEnvFileValues_MbcliYAMLNestedVaultNoEnvVault(t *testing.T) {
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
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "mbcli.yaml"))
	t.Setenv("MBCLI_PROJECT_ROOT", "")

	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("FOO=def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(tmp, "mbcli.yaml")
	y := "envs:\n  FOO: root\n  staging:\n    FOO: stg\n"
	if err := os.WriteFile(yamlPath, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "root" {
		t.Fatalf("FOO=%q want root (sem --env-vault não aplica sub-vault mbcli)", m["FOO"])
	}
}

func TestBuildEnvFileValues_CwdDotEnvOverMbcliYAML(t *testing.T) {
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
	t.Setenv("MBCLI_YAML_PATH", filepath.Join(tmp, "mbcli.yaml"))
	t.Setenv("MBCLI_PROJECT_ROOT", "")

	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("A=def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(tmp, "mbcli.yaml"),
		[]byte("envs:\n  A: from-yaml\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	dotEnv := filepath.Join(tmp, ".env")
	if err := os.WriteFile(dotEnv, []byte("A=cwd\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
	}
	m, err := BuildEnvFileValues(rt, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m["A"] != "cwd" {
		t.Fatalf("A=%q, want cwd (.env por cima de mbcli.yaml)", m["A"])
	}
}

func TestBuildEnvFileValues_ResolvesOPReferences(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("# empty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(def+".secrets", []byte("TOKEN\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	secrets := stringSecretStore{}
	secrets.Set("default", "TOKEN", "op://Private/item-1/TOKEN")
	op := stubOnePassword{
		read: func(ref string) (string, error) {
			if ref == "op://Private/item-1/TOKEN" {
				return "from-1p", nil
			}
			return "", errors.New("unexpected ref")
		},
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
	}
	m, err := BuildEnvFileValues(rt, secrets, op)
	if err != nil {
		t.Fatal(err)
	}
	if m["TOKEN"] != "from-1p" {
		t.Fatalf("TOKEN=%q", m["TOKEN"])
	}
}

func TestBuildEnvFileValues_OPReferenceWithoutReaderErrors(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(def+".secrets", []byte("X\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	secrets := stringSecretStore{}
	secrets.Set("default", "X", "op://v/i/X")
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
	}
	_, err := BuildEnvFileValues(rt, secrets, nil)
	if err == nil {
		t.Fatal("expected error when op:// present and OnePassword is nil")
	}
}

func TestBuildEnvFileValues_InvalidEnvVault(t *testing.T) {
	tmp := t.TempDir()
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: filepath.Join(tmp, "env.defaults"),
		},
		EnvVault: "../x",
	}
	_, err := BuildEnvFileValues(rt, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildEnvFileValues_EnvVaultProjectReserved(t *testing.T) {
	tmp := t.TempDir()
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: filepath.Join(tmp, "env.defaults"),
		},
		EnvVault: "project",
	}
	_, err := BuildEnvFileValues(rt, nil, nil)
	if err == nil {
		t.Fatal("expected error for reserved EnvVault project")
	}
}

func TestBuildEnvFileValues_OPSecretsFile(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("#\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(def+".opsecrets", []byte(`TOKEN=op://Private/item-1/TOKEN
`), 0o644); err != nil {
		t.Fatal(err)
	}
	op := stubOnePassword{
		read: func(ref string) (string, error) {
			if ref == "op://Private/item-1/TOKEN" {
				return "from-file-op", nil
			}
			return "", errors.New("unexpected ref")
		},
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
		},
	}
	m, err := BuildEnvFileValues(rt, nil, op)
	if err != nil {
		t.Fatal(err)
	}
	if m["TOKEN"] != "from-file-op" {
		t.Fatalf("TOKEN=%q", m["TOKEN"])
	}
}
