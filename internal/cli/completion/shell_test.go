package completion

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizeShellName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"bash", ShellBash},
		{"/bin/bash", ShellBash},
		{"/usr/local/bin/zsh", ShellZsh},
		{"-zsh", ShellZsh},
		{"pwsh", ShellPowerShell},
		{"powershell.exe", ShellPowerShell},
		{"fish", ShellFish},
	}
	for _, tc := range tests {
		got, err := NormalizeShellName(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("NormalizeShellName(%q) = %q, %v; want %q, nil", tc.in, got, err, tc.want)
		}
	}
	if _, err := NormalizeShellName("csh"); err == nil {
		t.Fatal("expected error for csh")
	}
}

func TestProfilePath_override(t *testing.T) {
	p := filepath.Join(t.TempDir(), "customrc")
	got, err := ProfilePath(ShellZsh, p)
	if err != nil || got != p {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestProfilePath_defaults(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))

	paths := map[string]string{
		ShellBash:       filepath.Join(tmp, ".bashrc"),
		ShellZsh:        filepath.Join(tmp, ".zshrc"),
		ShellFish:       filepath.Join(tmp, "xdg", "fish", "config.fish"),
		ShellPowerShell: powerShellProfilePath(tmp),
	}
	for shell, want := range paths {
		got, err := ProfilePath(shell, "")
		if err != nil {
			t.Fatalf("%s: %v", shell, err)
		}
		if got != want {
			t.Errorf("shell %s: got %q want %q", shell, got, want)
		}
	}
}

func TestPowerShellProfilePath_windowsShape(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only path shape")
	}
	h := `C:\Users\test`
	got := powerShellProfilePath(h)
	if !strings.Contains(got, "Microsoft.PowerShell_profile.ps1") {
		t.Fatal(got)
	}
}
