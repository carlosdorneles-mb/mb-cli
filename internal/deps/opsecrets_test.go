package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOPSecretsRoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "env.defaults")
	if err := SetOPSecretRef(p, "K", "op://x/y/z"); err != nil {
		t.Fatal(err)
	}
	m, err := LoadOPSecretRefs(p)
	if err != nil {
		t.Fatal(err)
	}
	if m["K"] != "op://x/y/z" {
		t.Fatalf("got %#v", m)
	}
	if err := RemoveOPSecretRef(p, "K"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p + opSecretsSuffix); !os.IsNotExist(err) {
		t.Fatal("expected opsecrets file removed")
	}
}
