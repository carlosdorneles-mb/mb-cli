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

// User-facing copy for non-release builds (Portuguese CLI locale).
const cliUpdateNonReleaseMsg = `Este binário não veio da release oficial do MB CLI (build local, go install, etc.).
O comando mb update --only-cli só atualiza binários instalados a partir do GitHub Releases (versão embutida no executável).

Para instalar ou atualizar a versão estável:
  curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash

Releases: https://github.com/carlosdorneles-mb/mb-cli/releases
`

// RunCLIUpdate downloads and installs the latest MB CLI release when the binary is a release build
// (embedded version via ldflags). Non-release builds log cliUpdateNonReleaseMsg and return nil.
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

	if cfgDir := deps.Runtime.ConfigDir; cfgDir != "" {
		if _, shErr := shellhelpers.EnsureShellHelpers(cfgDir); shErr != nil {
			_ = log.Info(ctx, "Erro ao atualizar shell helpers: %v", shErr)
		}
	}

	return err
}
