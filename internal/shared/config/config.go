// Package config loads CLI application settings from ~/.config/mb/config.yaml.
// Plugin/command environment variables stay in .env.* files, not here.
package config

// DefaultDocsURL is the documentation site opened by mb --doc when docs_url is not set in config.yaml.
const DefaultDocsURL = "https://carlosdorneles-mb.github.io/mb-cli/"

// DefaultUpdateRepo is the default GitHub owner/repo for mb self update when not set in config or build.
const DefaultUpdateRepo = "carlosdorneles-mb/mb-cli"

// AppConfig holds effective CLI settings after loading config.yaml.
type AppConfig struct {
	// DocsBaseURL is opened by mb --doc.
	DocsBaseURL string
	// UpdateRepo is owner/repo for GitHub releases (mb self update).
	UpdateRepo string
}

// fileConfig is the struct used when deserializing config.yaml (yaml.Unmarshal).
// Pointer fields distinguish "key absent" (nil) from "key present" (non-nil).
// Validated with go-playground/validator/v10; see validator.go.
type fileConfig struct {
	DocsURL    *string `yaml:"docs_url"    validate:"omitempty,http_url"`
	UpdateRepo *string `yaml:"update_repo" validate:"omitempty,update_repo"`
}
