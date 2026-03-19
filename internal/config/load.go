package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"mb/internal/version"
)

// Load ensures config.yaml exists (creates it with defaults if missing), then reads and validates it.
// Precedence update_repo: key in file → version.UpdateRepo → DefaultUpdateRepo.
// Precedence docs_url: key in file (valid) → DefaultDocsURL.
func Load(configDir string) (AppConfig, error) {
	path := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return AppConfig{}, err
		}
		if err := writeDefaultConfigFile(configDir, path); err != nil {
			return AppConfig{}, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}

	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return AppConfig{}, fmt.Errorf("config.yaml: %w", err)
	}

	trimFileConfig(&fc)
	if err := validateFileConfig(&fc); err != nil {
		return AppConfig{}, fmt.Errorf("config.yaml: %w", err)
	}

	docsURL := DefaultDocsURL
	if fc.DocsURL != nil && strings.TrimSpace(*fc.DocsURL) != "" {
		docsURL = strings.TrimSpace(*fc.DocsURL)
	}

	updateRepo := updateRepoWithoutFile()
	if fc.UpdateRepo != nil && strings.TrimSpace(*fc.UpdateRepo) != "" {
		updateRepo = strings.TrimSpace(*fc.UpdateRepo)
	}

	return AppConfig{DocsBaseURL: docsURL, UpdateRepo: updateRepo}, nil
}

func trimFileConfig(fc *fileConfig) {
	if fc.DocsURL != nil {
		s := strings.TrimSpace(*fc.DocsURL)
		fc.DocsURL = &s
	}
	if fc.UpdateRepo != nil {
		s := strings.TrimSpace(*fc.UpdateRepo)
		fc.UpdateRepo = &s
	}
}

func updateRepoWithoutFile() string {
	if r := strings.TrimSpace(version.UpdateRepo); r != "" {
		return r
	}
	return DefaultUpdateRepo
}
