package shell

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all.sh log.sh
var shellFS embed.FS

// EnsureShellHelpers cria o diretório lib/shell sob configDir (se não existir),
// grava all.sh e log.sh a partir do conteúdo embutido, e retorna o path
// absoluto do diretório lib/shell. Esse path é passado ao plugin como MB_HELPERS_PATH.
func EnsureShellHelpers(configDir string) (string, error) {
	shellDir := filepath.Join(configDir, "lib", "shell")
	allPath := filepath.Join(shellDir, "all.sh")
	if _, err := os.Stat(allPath); err == nil {
		return filepath.Abs(shellDir)
	}
	if err := os.MkdirAll(shellDir, 0o755); err != nil {
		return "", err
	}
	for _, name := range []string{"all.sh", "log.sh"} {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(shellDir, name), data, 0o644); err != nil {
			return "", err
		}
	}
	return filepath.Abs(shellDir)
}
