package alias

import (
	"sort"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
)

func sortAliasStoreKeys(keys []string) {
	sort.Slice(keys, func(i, j int) bool {
		vi, ni, okI := alib.ParseStoreKey(keys[i])
		vj, nj, okJ := alib.ParseStoreKey(keys[j])
		if !okI || !okJ {
			return keys[i] < keys[j]
		}
		if ni != nj {
			return ni < nj
		}
		return vi < vj
	})
}

// aliasListRow is one merged row for mb alias list (project overlays global for same store key).
type aliasListRow struct {
	Name      string
	EnvVault  string
	Command   []string
	Source    string // "config" | "project"
	MbcliPath string // absolute mbcli.yaml when Source is project
}

func loadMergedAliasRows(cfgDir string) ([]aliasListRow, error) {
	path := alib.FilePath(cfgDir)
	global, err := alib.Load(path)
	if err != nil {
		return nil, err
	}
	mbcliPath, err := deps.ResolveMbcliYAMLPath()
	if err != nil {
		return nil, err
	}
	project, err := deps.ParseMbcliAliases(mbcliPath)
	if err != nil {
		return nil, err
	}
	keySet := make(map[string]struct{})
	for k := range global.Aliases {
		keySet[k] = struct{}{}
	}
	for k := range project {
		keySet[k] = struct{}{}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sortAliasStoreKeys(keys)
	out := make([]aliasListRow, 0, len(keys))
	for _, sk := range keys {
		vault, name, ok := alib.ParseStoreKey(sk)
		if !ok {
			continue
		}
		if e, ok := project[sk]; ok {
			out = append(out, aliasListRow{
				Name:      name,
				EnvVault:  vault,
				Command:   append([]string(nil), e.Command...),
				Source:    "project",
				MbcliPath: mbcliPath,
			})
			continue
		}
		e := global.Aliases[sk]
		out = append(out, aliasListRow{
			Name:      name,
			EnvVault:  vault,
			Command:   append([]string(nil), e.Command...),
			Source:    "config",
			MbcliPath: "",
		})
	}
	return out, nil
}
