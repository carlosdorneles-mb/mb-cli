package plugins

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FlagDef defines a flag that can run an entrypoint (for manifests without default entrypoint).
// Used internally and in FlagsJSON; produced from FlagEntry (list format in manifest).
type FlagDef struct {
	Type        string   `yaml:"type"        json:"type"`        // "long" (e.g. --flag) or "short" (-s)
	Short       string   `yaml:"short"       json:"short"`       // optional; single letter for -x form
	Entrypoint  string   `yaml:"entrypoint"  json:"entrypoint"`  // script to run when this flag is set
	Description string   `yaml:"description" json:"description"` // shown in command help (Cobra usage)
	Envs        []string `yaml:"envs"        json:"envs"`        // optional; KEY=VALUE pairs injected only when flag is used
}

// FlagEntry is one item in the list format of flags in the manifest (name, description, entrypoint, commands).
type FlagEntry struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Entrypoint  string   `yaml:"entrypoint"`
	Envs        []string `yaml:"envs"`
	Commands    struct {
		Long  string `yaml:"long"`
		Short string `yaml:"short"`
	} `yaml:"commands"`
}

// FlagsSpec holds the list of flag entries (name, description, entrypoint, commands) in the manifest.
type FlagsSpec struct {
	List []FlagEntry
}

// UnmarshalYAML accepts only a list of FlagEntry; map format is no longer supported.
func (f *FlagsSpec) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf(
			"flags must be a list of flag entries (name, description, entrypoint, commands)",
		)
	}
	f.List = nil
	return value.Decode(&f.List)
}

// Len returns the number of flags.
func (f *FlagsSpec) Len() int {
	return len(f.List)
}

// ToMap returns a map[string]FlagDef for use by the scanner (FlagsJSON) and dynamic commands.
// Key = commands.long or name; value = FlagDef with Short, Entrypoint, Description.
func (f *FlagsSpec) ToMap() map[string]FlagDef {
	if len(f.List) == 0 {
		return nil
	}
	m := make(map[string]FlagDef)
	for _, e := range f.List {
		key := e.Commands.Long
		if key == "" {
			key = e.Name
		}
		m[key] = FlagDef{
			Type:        "long",
			Short:       e.Commands.Short,
			Entrypoint:  e.Entrypoint,
			Description: e.Description,
			Envs:        append([]string(nil), e.Envs...),
		}
	}
	return m
}

// Manifest is the per-directory manifest. Path hierarchy defines categories/subcategories.
type Manifest struct {
	Command         string       `yaml:"command"`          // optional; default = dir name
	Description     string       `yaml:"description"`      // Short for Cobra
	LongDescription string       `yaml:"long_description"` // optional; Long for Cobra (multi-line ok)
	Entrypoint      string       `yaml:"entrypoint"`       // if set = executable leaf (type inferred by .sh suffix)
	Readme          string       `yaml:"readme"`
	Flags           FlagsSpec    `yaml:"flags"`      // list only; flags-only leaf when Entrypoint is empty; optional extra scripts when Entrypoint is set
	Use             string       `yaml:"use"`        // optional; Cobra Use template (e.g. "<name>" or "[env]")
	Args            int          `yaml:"args"`       // optional; number of required positional args (0 = no validation)
	Aliases         []string     `yaml:"aliases"`    // optional; Cobra Aliases
	Example         string       `yaml:"example"`    // optional; Cobra Example
	Deprecated      string       `yaml:"deprecated"` // optional; Cobra Deprecated message
	Hidden          bool         `yaml:"hidden"`     // optional; omit from mb help (still invokable)
	EnvFiles        EnvFilesSpec `yaml:"env_files"`  // optional; .env per group (executável only)
	GroupID         string       `yaml:"group_id"`   // optional; nested leaves only, see groups.yaml
}

// ManifestEnvGroupDefault is the logical group when env_files entry omits group or --env-group is not set.
const ManifestEnvGroupDefault = "default"

// EnvFileEntry is one manifest env_files item (paths relative to plugin dir).
type EnvFileEntry struct {
	File  string `yaml:"file"  json:"file"`
	Group string `yaml:"group" json:"group"`
}

// EnvFilesSpec is a list of env file bindings in the manifest.
type EnvFilesSpec struct {
	List []EnvFileEntry
}

// UnmarshalYAML accepts only a sequence of { file, group }.
func (e *EnvFilesSpec) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("env_files must be a list of entries (file, optional group)")
	}
	e.List = nil
	return value.Decode(&e.List)
}

// Len returns the number of env_files entries.
func (e *EnvFilesSpec) Len() int {
	return len(e.List)
}

func (m *Manifest) normalizeEnvFileGroups() {
	for i := range m.EnvFiles.List {
		if strings.TrimSpace(m.EnvFiles.List[i].Group) == "" {
			m.EnvFiles.List[i].Group = ManifestEnvGroupDefault
		} else {
			m.EnvFiles.List[i].Group = strings.TrimSpace(m.EnvFiles.List[i].Group)
		}
	}
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
	manifest.normalizeEnvFileGroups()

	return manifest, raw, nil
}
