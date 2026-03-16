package plugins

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FlagDef defines a flag that can run an entrypoint (for manifests without default entrypoint).
type FlagDef struct {
	Type       string `yaml:"type"`        // "long" (--name) or "short" (-s)
	Short      string `yaml:"short"`      // optional; single letter for -x form; used with long name = map key
	Entrypoint string `yaml:"entrypoint"`  // script to run when this flag is set
}

// Manifest is the per-directory manifest. Path hierarchy defines categories/subcategories.
type Manifest struct {
	Command     string             `yaml:"command"`     // optional; default = dir name
	Description string             `yaml:"description"` // Short for Cobra
	Entrypoint  string             `yaml:"entrypoint"`  // if set = executable leaf (type inferred by .sh suffix)
	Readme      string             `yaml:"readme"`
	Flags       map[string]FlagDef `yaml:"flags"` // only used when Entrypoint is empty (flags-only leaf)
}

// PluginTypeFromEntrypoint returns "sh" if entrypoint ends with .sh, otherwise "bin".
func PluginTypeFromEntrypoint(entrypoint string) string {
	if strings.HasSuffix(entrypoint, ".sh") {
		return "sh"
	}
	return "bin"
}

func LoadManifest(path string) (Manifest, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, nil, err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return Manifest{}, nil, err
	}

	return manifest, raw, nil
}
