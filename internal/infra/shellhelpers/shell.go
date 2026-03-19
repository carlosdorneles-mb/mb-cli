package shellhelpers

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed *.sh
var shellFS embed.FS

const checksumFile = ".checksum"

// embeddedShellFiles retorna os nomes dos arquivos .sh embutidos, ordenados para checksum determinístico.
func embeddedShellFiles() ([]string, error) {
	entries, err := fs.Glob(shellFS, "*.sh")
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	return entries, nil
}

// helpersChecksum retorna SHA256 (hex) da concatenação do conteúdo dos arquivos .sh embutidos.
func helpersChecksum() (string, error) {
	files, err := embeddedShellFiles()
	if err != nil {
		return "", err
	}
	h := sha256.New()
	for _, name := range files {
		data, err := fs.ReadFile(shellFS, name)
		if err != nil {
			return "", err
		}
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnsureShellHelpers cria o diretório lib/shell sob configDir (se não existir),
// grava todos os .sh embutidos (descobertos automaticamente) e retorna o path
// absoluto do diretório lib/shell. Esse path é passado ao plugin como MB_HELPERS_PATH.
// Se o arquivo .checksum em lib/shell existir e for igual ao checksum atual do embed,
// os arquivos não são reescritos (atualização automática quando o conteúdo muda).
func EnsureShellHelpers(configDir string) (string, error) {
	shellDir := filepath.Join(configDir, "lib", "shell")
	files, err := embeddedShellFiles()
	if err != nil {
		return "", err
	}
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

	for _, name := range files {
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
