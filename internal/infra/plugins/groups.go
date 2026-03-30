package plugins

import (
	"fmt"
	"os"

	"mb/internal/domain/plugin"
)

// LoadGroupsFile reads and validates groups.yaml. Missing file returns nil, nil.
func LoadGroupsFile(path string) ([]plugin.HelpGroupDef, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	list, err := plugin.ParseHelpGroupsYAML(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return list, nil
}
