package update

import (
	"context"
	"strings"

	"mb/internal/deps"
	"mb/internal/infra/selfupdate"
	"mb/internal/shared/system"
)

// LogInfoLines logs each non-empty line of text at info level.
func LogInfoLines(ctx context.Context, log *system.Logger, text string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			_ = log.Info(ctx, "%s", line)
		}
	}
}

// LogCheckOnlyHumanLines logs CheckOnly text: info for headlines, Print (gum none) for version lines.
func LogCheckOnlyHumanLines(ctx context.Context, log *system.Logger, text string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if selfupdate.IsPlainCheckOnlyLine(line) {
			_ = log.Print(ctx, "%s", line)
			continue
		}
		_ = log.Info(ctx, "%s", line)
	}
}

// SelfUpdateConfigFromDeps builds selfupdate.Config from application config (UpdateRepo override).
func SelfUpdateConfigFromDeps(d deps.Dependencies) *selfupdate.Config {
	cfg := &selfupdate.Config{}
	if r := strings.TrimSpace(d.AppConfig.UpdateRepo); r != "" {
		cfg.Repo = r
	}
	return cfg
}
