package update

import (
	"context"
	"strings"

	"mb/internal/deps"
	"mb/internal/infra/selfupdate"
	"mb/internal/infra/shellhelpers"
	"mb/internal/shared/system"
	"mb/internal/shared/version"
)

const cliUpdateNonReleaseMsg = `Este binário não veio da release oficial do MB CLI (build local, go install, etc.).
O commando mb update --only-cli só atualiza binários instalados a partir do GitHub Releases (versão embutida no executável).

Para instalar ou atualizar a versão estável:
  curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash

Releases: https://github.com/carlosdorneles-mb/mb-cli/releases
`

func logInfoLines(ctx context.Context, log *system.Logger, text string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			_ = log.Info(ctx, "%s", line)
		}
	}
}

// RunCLIUpdate updates the MB CLI binary to the latest release. Non-release builds get the usual message and nil.
func RunCLIUpdate(ctx context.Context, deps deps.Dependencies, log *system.Logger) error {
	_ = log.Info(ctx, "Atualizando MB CLI...")

	if !version.IsReleaseBuild() {
		logInfoLines(ctx, log, cliUpdateNonReleaseMsg)
		return nil
	}

	local := strings.TrimSpace(version.Version)
	suCfg := selfupdateFromAppConfig(deps)
	out, err := selfupdate.Run(ctx, suCfg, local)
	if out != "" {
		logInfoLines(ctx, log, out)
	}

	// Ensure shell helpers are updated after CLI update
	if cfgDir := deps.Runtime.ConfigDir; cfgDir != "" {
		if _, shErr := shellhelpers.EnsureShellHelpers(cfgDir); shErr != nil {
			_ = log.Info(ctx, "Erro ao atualizar shell helpers: %v", shErr)
		}
	}

	return err
}

func selfupdateFromAppConfig(d deps.Dependencies) *selfupdate.Config {
	cfg := &selfupdate.Config{}
	if r := strings.TrimSpace(d.AppConfig.UpdateRepo); r != "" {
		cfg.Repo = r
	}
	return cfg
}
