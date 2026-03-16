package shell

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed index.sh log.sh
var shellFS embed.FS

// EnsureShellHelpers cria o diretório lib/shell sob configDir (se não existir),
// grava index.sh e log.sh a partir do conteúdo embutido, e retorna o path
// absoluto do index.sh. Esse path deve ser passado ao plugin como MB_HELPERS_PATH.
func EnsureShellHelpers(configDir string) (string, error) {
	shellDir := filepath.Join(configDir, "lib", "shell")
	indexPath := filepath.Join(shellDir, "index.sh")
	if _, err := os.Stat(indexPath); err == nil {
		return indexPath, nil
	}
	if err := os.MkdirAll(shellDir, 0o755); err != nil {
		return "", err
	}
	for _, name := range []string{"index.sh", "log.sh"} {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(shellDir, name), data, 0o644); err != nil {
			return "", err
		}
	}
	return indexPath, nil
}
