package env

import (
	"sort"

	"mb/internal/deps"
)

func envTargetPath(d deps.Dependencies, group string) (string, error) {
	if group == "" {
		return d.Runtime.DefaultEnvPath, nil
	}
	if err := deps.ValidateEnvGroup(group); err != nil {
		return "", err
	}
	return deps.GroupEnvFilePath(d.Runtime.ConfigDir, group)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
