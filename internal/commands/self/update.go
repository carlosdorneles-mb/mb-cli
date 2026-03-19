package self

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/selfupdate"
	"mb/internal/system"
	"mb/internal/version"
)

const selfUpdateNonReleaseMsg = `Este binário não veio da release oficial do MB CLI (build local, go install, etc.).
O commando mb self update só atualiza binários instalados a partir do GitHub Releases (versão embutida no executável).

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

func newSelfUpdateCmd(deps deps.Dependencies) *cobra.Command {
	var checkOnly bool
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"up", "u"},
		Short:   "Atualiza o MB CLI para a última release estável do GitHub",
		Long: `Só se aplica a binários da release oficial (versão definida no build). Builds locais recebem uma mensagem a orientar o install.sh.

Consulta o GitHub, compara com a versão embutida e, se houver release mais nova, baixa o binário mb (SHA256) e substitui o executável.

Só atualiza o binário mb (não reinstala gum, glow, jq nem fzf). Linux/macOS, amd64/arm64.

Com --check-only apenas verifica se há release mais nova (sem download). Códigos de saída: 0 = ok; 2 = há atualização; 1 = erro.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			quiet := deps.Runtime != nil && deps.Runtime.Quiet
			log := system.NewLogger(
				quiet,
				deps.Runtime != nil && deps.Runtime.Verbose,
				cmd.ErrOrStderr(),
			)
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if !version.IsReleaseBuild() {
				if !quiet {
					logInfoLines(ctx, log, selfUpdateNonReleaseMsg)
				}
				return nil
			}
			local := strings.TrimSpace(version.Version)
			suCfg := selfupdateFromAppConfig(deps)
			if checkOnly {
				out, code, err := selfupdate.RunCheckOnly(ctx, suCfg, local)
				if out != "" && !quiet {
					logInfoLines(ctx, log, out)
				}
				if err != nil {
					return err
				}
				if code == selfupdate.ExitCodeUpdateAvailable {
					os.Exit(selfupdate.ExitCodeUpdateAvailable)
				}
				return nil
			}
			out, err := selfupdate.Run(ctx, suCfg, local)
			if out != "" && !quiet {
				logInfoLines(ctx, log, out)
			}
			return err
		},
	}
	cmd.Flags().
		BoolVar(&checkOnly, "check-only", false, "Só verifica se há atualização (sem baixar)")
	return cmd
}

func selfupdateFromAppConfig(d deps.Dependencies) *selfupdate.Config {
	cfg := &selfupdate.Config{}
	if r := strings.TrimSpace(d.AppConfig.UpdateRepo); r != "" {
		cfg.Repo = r
	}
	return cfg
}
