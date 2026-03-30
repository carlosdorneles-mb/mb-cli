package plugins

import (
	"fmt"
	"path/filepath"

	"mb/internal/ports"
)

func dirExistsFS(fsys ports.Filesystem, p string) bool {
	info, err := fsys.Stat(p)
	return err == nil && info.IsDir()
}

func uniqueInstallDir(fsys ports.Filesystem, pluginsDir, base string) string {
	for i := 2; ; i++ {
		installDir := fmt.Sprintf("%s-%d", base, i)
		destDir := filepath.Join(pluginsDir, installDir)
		if !dirExistsFS(fsys, destDir) {
			return installDir
		}
	}
}
