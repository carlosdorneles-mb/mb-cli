package alias

import (
	"fmt"
	"path/filepath"
	"strings"

	"mb/internal/cli/completion"
	alib "mb/internal/shared/aliases"
)

// profileSourceLine returns the line to append inside the profile block for the given shell.
func profileSourceLine(configDir, shell string) (string, error) {
	shellDir := alib.ShellDir(configDir)
	var rel string
	switch shell {
	case completion.ShellBash, completion.ShellZsh:
		rel = filepath.Join(shellDir, "aliases.bash")
	case completion.ShellFish:
		rel = filepath.Join(shellDir, "aliases.fish")
	case completion.ShellPowerShell:
		rel = filepath.Join(shellDir, "aliases.ps1")
	default:
		return "", fmt.Errorf("shell desconhecido: %s", shell)
	}
	abs, err := filepath.Abs(rel)
	if err != nil {
		return "", err
	}
	switch shell {
	case completion.ShellBash, completion.ShellZsh:
		return fmt.Sprintf(". %s", posixSQ(abs)), nil
	case completion.ShellFish:
		return fmt.Sprintf("source %s", posixSQ(abs)), nil
	case completion.ShellPowerShell:
		return fmt.Sprintf(". %s", ps1SQ(abs)), nil
	default:
		return "", fmt.Errorf("shell desconhecido: %s", shell)
	}
}

func posixSQ(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `'\''`) + `'`
}

func ps1SQ(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
