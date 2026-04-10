package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"mb/internal/shared/safepath"
)

// MergeManifestEnvFiles loads KEY=VALUE from manifest env_files entries whose vault
// matches effectiveVault (use ManifestEnvVaultDefault when --env-vault is unset).
// Later files in the list override earlier ones for duplicate keys.
func MergeManifestEnvFiles(
	pluginDir, envFilesJSON, effectiveVault string,
) (map[string]string, error) {
	if envFilesJSON == "" || pluginDir == "" {
		return map[string]string{}, nil
	}
	var entries []EnvFileEntry
	if err := json.Unmarshal([]byte(envFilesJSON), &entries); err != nil {
		return nil, fmt.Errorf("env_files do plugin: JSON inválido: %w", err)
	}
	merged := map[string]string{}
	for _, e := range entries {
		if e.Vault != effectiveVault {
			continue
		}
		abs := filepath.Join(pluginDir, e.File)
		if err := safepath.ValidateUnderDir(abs, pluginDir); err != nil {
			return nil, fmt.Errorf("env_files %q: path inválido: %w", e.File, err)
		}
		values, err := godotenv.Read(abs)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("arquivo de ambiente do plugin não encontrado: %s", e.File)
			}
			return nil, fmt.Errorf("ler env_files %q: %w", e.File, err)
		}
		for k, v := range values {
			merged[k] = v
		}
	}
	return merged, nil
}
