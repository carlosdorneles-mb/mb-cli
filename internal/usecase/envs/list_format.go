package envs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"mb/internal/shared/system"
)

// FormatJSONByVault writes the rows as a flat array with vault metadata.
func FormatJSONByVault(w io.Writer, rows []ListRow, showSecrets bool, configDir string) error {
	// Array plano ao invés de mapa agrupado
	entries := make([]map[string]interface{}, 0, len(rows))

	for _, r := range rows {
		entry := map[string]interface{}{
			"vault":   r.Vault,
			"key":     r.Key,
			"value":   r.Value,
			"storage": r.Storage,
			"path":    CalculateEnvFilePath(r.Vault, r.Storage, configDir),
		}

		// Detect if it's a secret
		isSecret := (r.Storage == StorageKeyring || r.Storage == Storage1Password)
		entry["isSecret"] = isSecret

		// Mask secrets unless showing secrets
		if isSecret && !showSecrets {
			entry["value"] = "***"
		}

		entries = append(entries, entry)
	}

	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

func formatTextPlain(w io.Writer, rows []ListRow) error {
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%s=%s\n", r.Key, r.Value); err != nil {
			return err
		}
	}
	return nil
}

func formatTable(ctx context.Context, w io.Writer, rows []ListRow, configDir string) error {
	table := make([][]string, len(rows))
	for i, r := range rows {
		// Calculate file path
		path := CalculateEnvFilePath(r.Vault, r.Storage, configDir)
		// Truncate long paths
		displayPath := path
		if len(displayPath) > 40 {
			displayPath = "..." + displayPath[len(displayPath)-37:]
		}

		table[i] = []string{
			r.Key + "=" + r.Value,
			r.Vault,
			r.Storage,
			displayPath,
		}
	}
	headers := []string{"VAR", "VAULT", "ARMAZENAMENTO", "ARQUIVO"}
	return system.GumTable(ctx, headers, table, w)
}

// CalculateEnvFilePath calculates the file path where a variable is stored
func CalculateEnvFilePath(vault, storage, configDir string) string {
	switch storage {
	case StorageKeyring:
		return filepath.Join(configDir, ".env."+vault+".secrets")
	case Storage1Password:
		return filepath.Join(configDir, ".env."+vault+".opsecrets")
	case StorageProject:
		return "mbcli.yaml"
	default:
		if vault == "default" || vault == "" {
			return filepath.Join(configDir, "env.defaults")
		}
		return filepath.Join(configDir, ".env."+vault)
	}
}
