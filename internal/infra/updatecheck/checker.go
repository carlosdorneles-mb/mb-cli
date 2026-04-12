// Package updatecheck handles automatic version checking for the MB CLI binary.
// It checks for new releases every hour and stores the result in the config directory.
package updatecheck

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mb/internal/infra/selfupdate"
	"mb/internal/shared/version"
)

const (
	// CheckInterval is how often we query the GitHub API for new releases.
	CheckInterval = 1 * time.Hour

	// DisableEnvVar can be set to "1" or "true" to disable version checks.
	DisableEnvVar = "MB_DISABLE_UPDATE_CHECK"

	lastCheckFile = ".last-update-check"
	availableFile = ".update-available"
	checkTimeout  = 30 * time.Second
)

// Checker performs version checks against the configured GitHub repository.
type Checker struct {
	configDir      string
	currentVersion string
	httpClient     *http.Client
}

// NewChecker creates a Checker that stores state in configDir.
func NewChecker(configDir, currentVersion string) *Checker {
	return &Checker{
		configDir:      configDir,
		currentVersion: currentVersion,
		httpClient:     &http.Client{Timeout: checkTimeout},
	}
}

// IsDisabled returns true when the env var MB_DISABLE_UPDATE_CHECK is "1" or "true".
func IsDisabled() bool {
	v := strings.TrimSpace(os.Getenv(DisableEnvVar))
	return v == "1" || strings.EqualFold(v, "true")
}

// ShouldCheck reports whether enough time has passed since the last check.
func (c *Checker) ShouldCheck() bool {
	path := filepath.Join(c.configDir, lastCheckFile)
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return true // nunca verificou
	}
	lastCheck, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return true // arquivo corrompido
	}
	elapsed := time.Since(time.Unix(lastCheck, 0))
	return elapsed >= CheckInterval
}

// RecordCheck writes the current timestamp so we don't check again for CheckInterval.
func (c *Checker) RecordCheck() error {
	path := filepath.Join(c.configDir, lastCheckFile)
	now := time.Now().Unix()
	return os.WriteFile(path, []byte(strconv.FormatInt(now, 10)), 0o644)
}

// IsUpdateAvailable returns the latest tag and true if a newer version was recorded.
func (c *Checker) IsUpdateAvailable() (string, bool) {
	path := filepath.Join(c.configDir, availableFile)
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return "", false
	}
	tag := strings.TrimSpace(string(data))
	if tag == "" {
		return "", false
	}
	return tag, true
}

// RecordAvailable writes the tag to availableFile so warnings can be shown.
func (c *Checker) RecordAvailable(tag string) error {
	path := filepath.Join(c.configDir, availableFile)
	return os.WriteFile(path, []byte(tag), 0o644)
}

// ClearAvailable removes the availableFile so warnings stop appearing.
func (c *Checker) ClearAvailable() {
	_ = os.Remove(filepath.Join(c.configDir, availableFile))
}

// Run performs the version check if enough time has passed since the last check.
// It is non-blocking on errors — failures are silently ignored.
func (c *Checker) Run(ctx context.Context) error {
	if IsDisabled() {
		return nil
	}
	if !c.ShouldCheck() {
		return nil
	}

	// Build configuration for selfupdate
	cfg := &selfupdate.Config{
		HTTPClient: c.httpClient,
	}

	// Fetch latest tag from GitHub
	latestTag, err := selfupdate.FetchLatestTag(ctx, cfg)
	if err != nil {
		// Falha na API — ignora silenciosamente
		return nil
	}

	// Always record the check so we don't hammer the API
	_ = c.RecordCheck()

	// Compare versions
	if !selfupdate.ShouldFetchNewRelease(c.currentVersion, latestTag) {
		// Mesma versão ou mais recente — limpa aviso anterior
		c.ClearAvailable()
		return nil
	}

	// Nova versão disponível — grava para mostrar warning
	if err := c.RecordAvailable(latestTag); err != nil {
		return fmt.Errorf("updatecheck: gravar disponível: %w", err)
	}
	return nil
}

// IsReleaseBuild mirrors version.IsReleaseBuild for convenience.
func IsReleaseBuild() bool {
	return version.IsReleaseBuild()
}
