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
const CLIUpdateNonReleaseMsg = `Este binário não veio da release oficial do MB CLI (build local, go install, etc.).
O comando mb update --only-cli só atualiza binários instalados a partir do GitHub Releases (versão embutida no executável).

Para instalar ou atualizar a versão estável:
  curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash

Releases: https://github.com/carlosdorneles-mb/mb-cli/releases
`

// RunCLIUpdate downloads and installs the latest MB CLI release when the binary is a release build
// (embedded version via ldflags). Non-release builds log CLIUpdateNonReleaseMsg and return nil.
func RunCLIUpdate(ctx context.Context, d deps.Dependencies, log *system.Logger) error {
	_ = log.Info(ctx, "Atualizando MB CLI...")

	if !version.IsReleaseBuild() {
		LogInfoLines(ctx, log, CLIUpdateNonReleaseMsg)
		return nil
	}

	local := strings.TrimSpace(version.Version)
	suCfg := SelfUpdateConfigFromDeps(d)
	out, err := selfupdate.Run(ctx, suCfg, local)
	if out != "" {
		LogInfoLines(ctx, log, out)
	}

	if d.Runtime != nil {
		if cfgDir := d.Runtime.ConfigDir; cfgDir != "" {
			if _, shErr := shellhelpers.EnsureShellHelpers(cfgDir); shErr != nil {
				_ = log.Info(ctx, "Erro ao atualizar shell helpers: %v", shErr)
			}
		}
	}

	return err
}
