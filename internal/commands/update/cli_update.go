package update

import (
	"context"
	"strings"

	"mb/internal/deps"
	"mb/internal/selfupdate"
	"mb/internal/shared/system"
	"mb/internal/version"
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
	return err
}

func selfupdateFromAppConfig(d deps.Dependencies) *selfupdate.Config {
	cfg := &selfupdate.Config{}
	if r := strings.TrimSpace(d.AppConfig.UpdateRepo); r != "" {
		cfg.Repo = r
	}
	return cfg
}
