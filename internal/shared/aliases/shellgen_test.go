package aliases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateBashZsh(t *testing.T) {
	f := &File{
		Version: 1,
		Aliases: map[string]Entry{
			StoreKey("", "dev"): {Command: []string{"docker", "compose", "up"}},
		},
	}
	s, err := generateBashZsh(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "dev()") || !strings.Contains(s, "command") ||
		!strings.Contains(s, "'docker'") {
		t.Fatalf("unexpected bash:\n%s", s)
	}
}

func TestWriteShellScripts_empty(t *testing.T) {
	dir := t.TempDir()
	if err := Save(FilePath(dir), &File{Version: 1, Aliases: map[string]Entry{}}); err != nil {
		t.Fatal(err)
	}
	if err := WriteShellScripts(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"aliases.bash", "aliases.fish", "aliases.ps1"} {
		b, err := os.ReadFile(filepath.Join(ShellDir(dir), name))
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Fatalf("empty %s", name)
		}
	}
}
