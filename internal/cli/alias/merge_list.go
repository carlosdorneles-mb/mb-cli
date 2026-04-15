package alias

import (
	"sort"

	"mb/internal/deps"
	alib "mb/internal/shared/aliases"
)

// aliasListRow is one merged row for mb alias list (project overlays global for same name).
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
	nameSet := make(map[string]struct{})
	for k := range global.Aliases {
		nameSet[k] = struct{}{}
	}
	for k := range project {
		nameSet[k] = struct{}{}
	}
	names := make([]string, 0, len(nameSet))
	for k := range nameSet {
		names = append(names, k)
	}
	sort.Strings(names)
	out := make([]aliasListRow, 0, len(names))
	for _, name := range names {
		if e, ok := project[name]; ok {
			out = append(out, aliasListRow{
				Name:      name,
				EnvVault:  e.EnvVault,
				Command:   append([]string(nil), e.Command...),
				Source:    "project",
				MbcliPath: mbcliPath,
			})
			continue
		}
		e := global.Aliases[name]
		out = append(out, aliasListRow{
			Name:      name,
			EnvVault:  e.EnvVault,
			Command:   append([]string(nil), e.Command...),
			Source:    "config",
			MbcliPath: "",
		})
	}
	return out, nil
}
