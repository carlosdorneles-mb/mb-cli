package envs

import (
	"sort"

	"mb/internal/deps"
	appenvs "mb/internal/usecase/envs"
)

func envPaths(d deps.Dependencies) appenvs.Paths {
	return appenvs.Paths{
		DefaultEnvPath: d.Runtime.DefaultEnvPath,
		ConfigDir:      d.Runtime.ConfigDir,
	}
}

func envTargetPath(d deps.Dependencies, group string) (string, error) {
	return appenvs.TargetPath(envPaths(d), group)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
