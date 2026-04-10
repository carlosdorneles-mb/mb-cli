package envs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSetSecretFlags_Explicit(t *testing.T) {
	t.Parallel()
	a, o, err := ResolveSetSecretFlags(true, false)
	if err != nil || !a || o {
		t.Fatalf("got asSecret=%v secretOP=%v err=%v", a, o, err)
	}
	a, o, err = ResolveSetSecretFlags(false, true)
	if err != nil || a || !o {
		t.Fatalf("got asSecret=%v secretOP=%v err=%v", a, o, err)
	}
}

func TestResolveSetSecretFlags_FromEnv(t *testing.T) {
	t.Setenv(MBEnvsSecretStoreKey, "keyring")
	a, o, err := ResolveSetSecretFlags(false, false)
	if err != nil || !a || o {
		t.Fatalf("got asSecret=%v secretOP=%v err=%v", a, o, err)
	}
}

func TestResolveSetSecretFlags_FromFileOverlay(t *testing.T) {
	_ = os.Unsetenv(MBEnvsSecretStoreKey)
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	_ = os.WriteFile(def, []byte(MBEnvsSecretStoreKey+"=op\n"), 0o644)
	vf := filepath.Join(tmp, ".env.stg")
	_ = os.WriteFile(vf, []byte(MBEnvsSecretStoreKey+"=keyring\n"), 0o644)
	a, o, err := ResolveSetSecretFlags(false, false, def, vf)
	if err != nil || !a || o {
		t.Fatalf("vault file should win: asSecret=%v secretOP=%v err=%v", a, o, err)
	}
}

func TestResolveSetSecretFlags_Invalid(t *testing.T) {
	t.Setenv(MBEnvsSecretStoreKey, "nope")
	_, _, err := ResolveSetSecretFlags(false, false)
	if err == nil {
		t.Fatal("expected error")
	}
}
