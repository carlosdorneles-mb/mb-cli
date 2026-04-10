package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveMbcliYAMLPath returns the absolute path to the project mbcli.yaml, matching
// mbcli-yaml.sh: MBCLI_YAML_PATH (priority), else ${MBCLI_PROJECT_ROOT:-.}/mbcli.yaml
// with "." meaning cwd; relative paths are joined with the current working directory.
// The file may not exist; callers decide whether that is an error.
func ResolveMbcliYAMLPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("mbcli.yaml: %w", err)
	}

	var f string
	if p := strings.TrimSpace(os.Getenv("MBCLI_YAML_PATH")); p != "" {
		f = p
	} else {
		root := strings.TrimSpace(os.Getenv("MBCLI_PROJECT_ROOT"))
		if root == "" {
			root = "."
		}
		if root == "." {
			root = wd
		} else if !filepath.IsAbs(root) {
			root = filepath.Join(wd, root)
		}
		root = filepath.Clean(root)
		f = filepath.Join(root, "mbcli.yaml")
	}

	if !filepath.IsAbs(f) {
		f = filepath.Join(wd, f)
	}
	return filepath.Clean(f), nil
}
