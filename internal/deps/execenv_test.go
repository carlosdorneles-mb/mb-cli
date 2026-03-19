package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mb/internal/shared/config"
)

func TestBuildMergedOSEnviron_NoOverlay(t *testing.T) {
	tmp := t.TempDir()
	def := filepath.Join(tmp, "env.defaults")
	if err := os.WriteFile(def, []byte("MB_EXECENV_TEST=fromdefaults\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rt := &RuntimeConfig{
		Paths: Paths{
			ConfigDir:      tmp,
			DefaultEnvPath: def,
			PluginsDir:     filepath.Join(tmp, "plugins"),
		},
	}
	d := NewDependencies(rt, config.AppConfig{}, nil, nil, nil)
	merged, err := BuildMergedOSEnviron(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	var found string
	for _, e := range merged {
		if strings.HasPrefix(e, "MB_EXECENV_TEST=") {
			found = e
			break
		}
	}
	if found != "MB_EXECENV_TEST=fromdefaults" {
		t.Fatalf(
			"want MB_EXECENV_TEST=fromdefaults in environ, got %q among %d entries",
			found,
			len(merged),
		)
	}
}
