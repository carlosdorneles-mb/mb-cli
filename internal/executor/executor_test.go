package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"mb/internal/infra/sqlite"
)

func TestRunInjectsEnv(t *testing.T) {
	tmp := t.TempDir()
	outputFile := filepath.Join(tmp, "out.txt")
	scriptPath := filepath.Join(tmp, "env.sh")
	script := "#!/bin/sh\necho \"$MB_TOKEN\" > \"$1\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	ex := New()
	plugin := sqlite.Plugin{
		CommandPath: "test/env",
		CommandName: "env",
		ExecPath:    scriptPath,
		PluginType:  "sh",
	}

	err := ex.Run(
		context.Background(),
		plugin,
		[]string{outputFile},
		[]string{"MB_TOKEN=abc123"},
		tmp,
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	raw, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(raw) != "abc123\n" {
		t.Fatalf("expected injected env value, got %q", string(raw))
	}
}

func TestRunRejectsPathOutsideAllowedRoot(t *testing.T) {
	tmp := t.TempDir()
	allowedRoot := filepath.Join(tmp, "plugin")
	scriptOutside := filepath.Join(tmp, "other", "run.sh")
	if err := os.MkdirAll(filepath.Dir(scriptOutside), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptOutside, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	ex := New()
	plugin := sqlite.Plugin{
		ExecPath:   scriptOutside,
		PluginType: "sh",
	}
	err := ex.Run(context.Background(), plugin, nil, nil, allowedRoot)
	if err == nil {
		t.Fatal("Run expected to fail when ExecPath is outside allowedRoot")
	}
}
