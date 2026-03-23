package update

import (
	"context"
	"strings"

	"mb/internal/deps"
	"mb/internal/infra/selfupdate"
	"mb/internal/shared/system"
)

// logInfoLines logs each non-empty line of text at info level (used by RunCLIUpdate and by
// the --check-only branch in NewUpdateCmd).
func logInfoLines(ctx context.Context, log *system.Logger, text string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			_ = log.Info(ctx, "%s", line)
		}
	}
}

// selfupdateFromAppConfig builds selfupdate.Config from application config (UpdateRepo override).
func selfupdateFromAppConfig(d deps.Dependencies) *selfupdate.Config {
	cfg := &selfupdate.Config{}
	if r := strings.TrimSpace(d.AppConfig.UpdateRepo); r != "" {
		cfg.Repo = r
	}
	return cfg
}
