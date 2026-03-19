package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultConfigFileMode = 0o600

func writeDefaultConfigFile(configDir, path string) error {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("config dir: %w", err)
	}
	data := []byte(
		"# MB CLI — configuração exclusives do CLI.\n" +
			"# Documentação: https://carlosdorneles-mb.github.io/mb-cli/docs/technical-reference/cli-config\n" +
			"\n\n",
	)

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".config.yaml.*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, defaultConfigFileMode); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
