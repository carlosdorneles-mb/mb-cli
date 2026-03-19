package plugins

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReservedHelpGroupIDs are Cobra root group IDs; plugins cannot redefine them.
var ReservedHelpGroupIDs = map[string]struct{}{
	"commands":        {},
	"plugin_commands": {},
}

// HelpGroupDef is one entry in groups.yaml (help sections for nested plugin commands).
type HelpGroupDef struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

var helpGroupIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// LoadGroupsFile reads and validates groups.yaml. Missing file returns nil, nil.
func LoadGroupsFile(path string) ([]HelpGroupDef, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var list []HelpGroupDef
	if err := yaml.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("groups.yaml: %w", err)
	}
	seen := make(map[string]struct{})
	for i, g := range list {
		g.ID = strings.TrimSpace(g.ID)
		g.Title = strings.TrimSpace(g.Title)
		if g.ID == "" {
			return nil, fmt.Errorf("groups.yaml: item %d: id não pode ser vazio", i+1)
		}
		if !helpGroupIDPattern.MatchString(g.ID) {
			return nil, fmt.Errorf(
				"groups.yaml: id %q inválido (use letras, números e _; comece com letra)",
				g.ID,
			)
		}
		if _, reserved := ReservedHelpGroupIDs[g.ID]; reserved {
			return nil, fmt.Errorf("groups.yaml: id %q é reservado", g.ID)
		}
		if g.Title == "" {
			return nil, fmt.Errorf("groups.yaml: grupo %q: title não pode ser vazio", g.ID)
		}
		if _, dup := seen[g.ID]; dup {
			return nil, fmt.Errorf("groups.yaml: id %q duplicado", g.ID)
		}
		seen[g.ID] = struct{}{}
		list[i] = g
	}
	return list, nil
}

// MergeHelpGroupsGlobal merges batches in order (each batch is typically one groups.yaml).
// Last definition per id wins; if title changes from a previous value, onConflict is called (may be nil).
func MergeHelpGroupsGlobal(groups [][]HelpGroupDef, onConflict func(msg string)) []HelpGroupDef {
	byID := make(map[string]string)
	for _, batch := range groups {
		for _, g := range batch {
			if prev, ok := byID[g.ID]; ok && prev != g.Title && onConflict != nil {
				onConflict(fmt.Sprintf(
					"plugin help: id=%q duplicado: título %q substituído por %q (último vence)",
					g.ID, prev, g.Title))
			}
			byID[g.ID] = g.Title
		}
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]HelpGroupDef, 0, len(ids))
	for _, id := range ids {
		out = append(out, HelpGroupDef{ID: id, Title: byID[id]})
	}
	return out
}
