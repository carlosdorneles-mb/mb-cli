package shell

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all.sh log.sh
var shellFS embed.FS

// embedFiles ordem fixa para checksum e escrita.
var embedFiles = []string{"all.sh", "log.sh"}

const checksumFile = ".checksum"

// helpersChecksum retorna SHA256 (hex) da concatenação do conteúdo dos arquivos
// embutidos na ordem embedFiles.
func helpersChecksum() (string, error) {
	h := sha256.New()
	for _, name := range embedFiles {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnsureShellHelpers cria o diretório lib/shell sob configDir (se não existir),
// grava all.sh e log.sh a partir do conteúdo embutido, e retorna o path
// absoluto do diretório lib/shell. Esse path é passado ao plugin como MB_HELPERS_PATH.
// Se o arquivo .checksum em lib/shell existir e for igual ao checksum atual do embed,
// os arquivos não são reescritos (atualização automática quando o conteúdo muda).
func EnsureShellHelpers(configDir string) (string, error) {
	shellDir := filepath.Join(configDir, "lib", "shell")
	currentChecksum, err := helpersChecksum()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(shellDir, 0o755); err != nil {
		return "", err
	}

	checksumPath := filepath.Join(shellDir, checksumFile)
	if data, err := os.ReadFile(checksumPath); err == nil {
		if strings.TrimSpace(string(data)) == currentChecksum {
			return filepath.Abs(shellDir)
		}
	}

	for _, name := range embedFiles {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(shellDir, name), data, 0o644); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(checksumPath, []byte(currentChecksum+"\n"), 0o644); err != nil {
		return "", err
	}
	return filepath.Abs(shellDir)
}
