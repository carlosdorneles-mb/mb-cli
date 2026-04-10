package envs

import (
	"path/filepath"
	"sort"
	"strings"

	"mb/internal/shared/envgroup"
)

// GroupRow is one env file location shown by mb envs groups.
type GroupRow struct {
	Group string `json:"group"`
	Path  string `json:"path"`
}

// CollectEnvGroupRows returns the default env file plus every per-group .env.<grupo> under config.
func CollectEnvGroupRows(paths Paths) ([]GroupRow, error) {
	matches, err := filepath.Glob(filepath.Join(paths.ConfigDir, ".env.*"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var rest []GroupRow
	for _, path := range matches {
		if strings.HasSuffix(path, secretsSuffix) {
			continue
		}
		base := filepath.Base(path)
		if !strings.HasPrefix(base, ".env.") {
			continue
		}
		g := strings.TrimPrefix(base, ".env.")
		if g == "" || envgroup.Validate(g) != nil {
			continue
		}
		rest = append(rest, GroupRow{Group: g, Path: path})
	}
	sort.Slice(rest, func(i, j int) bool {
		if rest[i].Group != rest[j].Group {
			return rest[i].Group < rest[j].Group
		}
		return rest[i].Path < rest[j].Path
	})
	out := make([]GroupRow, 0, 1+len(rest))
	out = append(out, GroupRow{Group: "default", Path: paths.DefaultEnvPath})
	out = append(out, rest...)
	return out, nil
}
