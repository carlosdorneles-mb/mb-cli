package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Shell names suportados por Gen* do Cobra.
const (
	ShellBash       = "bash"
	ShellZsh        = "zsh"
	ShellFish       = "fish"
	ShellPowerShell = "powershell"
)

// NormalizeShellName converte basename de SHELL ou nome curto no shell canónico.
func NormalizeShellName(shell string) (string, error) {
	s := strings.TrimSpace(shell)
	if s == "" {
		return "", fmt.Errorf("shell vazio")
	}
	base := filepath.Base(s)
	base = strings.TrimSuffix(strings.ToLower(base), ".exe")
	switch base {
	case "bash", "rbash":
		return ShellBash, nil
	case "zsh", "-zsh":
		return ShellZsh, nil
	case "fish":
		return ShellFish, nil
	case "powershell", "pwsh":
		return ShellPowerShell, nil
	default:
		return "", fmt.Errorf("shell não suportado %q (use %s, %s, %s ou %s)",
			shell, ShellBash, ShellZsh, ShellFish, ShellPowerShell)
	}
}

// DetectShell usa SHELL em Unix; em Windows sem SHELL útil, assume PowerShell.
func DetectShell() (string, error) {
	sh := strings.TrimSpace(os.Getenv("SHELL"))
	if sh != "" {
		return NormalizeShellName(sh)
	}
	if runtime.GOOS == "windows" {
		return ShellPowerShell, nil
	}
	return "", fmt.Errorf("variável SHELL não definida; defina-a ou use --shell")
}

// ProfilePath devolve o ficheiro de perfil por omissão para o shell.
// rcFileOverride, se não vazio, tem precedência.
func ProfilePath(shell string, rcFileOverride string) (string, error) {
	if strings.TrimSpace(rcFileOverride) != "" {
		return filepath.Clean(rcFileOverride), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	switch shell {
	case ShellBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellZsh:
		return filepath.Join(home, ".zshrc"), nil
	case ShellFish:
		xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
		if xdg == "" {
			xdg = filepath.Join(home, ".config")
		}
		return filepath.Join(xdg, "fish", "config.fish"), nil
	case ShellPowerShell:
		return powerShellProfilePath(home), nil
	default:
		return "", fmt.Errorf("shell desconhecido: %s", shell)
	}
}

func powerShellProfilePath(home string) string {
	if runtime.GOOS == "windows" {
		ps7 := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		ps5dir := filepath.Join(home, "Documents", "WindowsPowerShell")
		ps5 := filepath.Join(ps5dir, "Microsoft.PowerShell_profile.ps1")
		if fi, err := os.Stat(filepath.Dir(ps7)); err == nil && fi.IsDir() {
			return ps7
		}
		return ps5
	}
	return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
}
