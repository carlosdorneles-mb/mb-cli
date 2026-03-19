package runtime

import (
	"mb/internal/deps"
	"mb/internal/shared/config"
)

// NewRuntimeConfig builds runtime config from resolved paths (CLI flags stay at zero values until Cobra runs).
func NewRuntimeConfig(p *deps.Paths) *deps.RuntimeConfig {
	return &deps.RuntimeConfig{Paths: *p}
}

// NewAppConfig loads ~/.config/mb/config.yaml (Viper + precedence for known keys).
func NewAppConfig(p *deps.Paths) (config.AppConfig, error) {
	return config.Load(p.ConfigDir)
}
