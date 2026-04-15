package system

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Regression: fzf preview scripts must not use echo "**...:** $(jq ...)" because bash
// re-expands $ inside the jq output while still inside double quotes.
func TestFzfPreviewBash_jqOutputNoDollarReexpand(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH")
	}

	type row struct {
		Name     string `json:"name"`
		Command  string `json:"command"`
		EnvVault string `json:"envVault"`
	}
	entries := []row{{
		Name:     "dudu2",
		Command:  "echo aqui $KEY",
		EnvVault: "staging",
	}}
	dir := t.TempDir()
	entriesPath := filepath.Join(dir, "entries.json")
	b, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entriesPath, b, 0o644); err != nil {
		t.Fatal(err)
	}

	// Mirrors the safe pattern in FzfTableWithPreviewForAliases (vars + printf).
	script := filepath.Join(dir, "preview_fragment.sh")
	fragment := `#!/bin/bash
set -euo pipefail
ENTRIES_FILE="$1"
INDEX=0
ENTRY=$(cat "$ENTRIES_FILE" | jq ".[$INDEX]")
export KEY=wrong
VAULT_RAW=$(echo "$ENTRY" | jq -r '.envVault // ""')
if [ -z "$VAULT_RAW" ]; then
  VAULT_LABEL="(nenhum)"
else
  VAULT_LABEL="$VAULT_RAW"
fi
NAME=$(echo "$ENTRY" | jq -r '.name // "N/A"')
CMD=$(echo "$ENTRY" | jq -r '.command // "N/A"')
OUT=$( {
  printf '%s\n' "### ${NAME}"
  echo ""
  printf '%s\n' "**Vault (mb run):** ${VAULT_LABEL}"
  printf '%s\n' "**Comando:** ${CMD}"
} )
if ! printf '%s' "$OUT" | grep -Fq '$KEY'; then
  printf 'expected literal $KEY in output, got:\n%s\n' "$OUT" >&2
  exit 1
fi
if printf '%s' "$OUT" | grep -Fq 'wrong'; then
  printf 'did not expect expanded KEY=wrong in output, got:\n%s\n' "$OUT" >&2
  exit 1
fi
`
	if err := os.WriteFile(script, []byte(fragment), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("bash", script, entriesPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bash preview fragment: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("unexpected stdout: %q", out)
	}
}
