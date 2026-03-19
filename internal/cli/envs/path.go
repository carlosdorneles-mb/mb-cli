package envs

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

// envGroupForKeyring returns the keyring group id: "" -> "default", else unchanged.
func envGroupForKeyring(group string) string {
	if group == "" {
		return "default"
	}
	return group
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
