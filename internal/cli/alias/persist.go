package alias

import alib "mb/internal/shared/aliases"

func saveAndRegenerate(configDir string, f *alib.File) error {
	if err := alib.Save(alib.FilePath(configDir), f); err != nil {
		return err
	}
	return alib.WriteShellScripts(configDir)
}
