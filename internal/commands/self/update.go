package self

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/selfupdate"
	"mb/internal/version"
)

const selfUpdateNonReleaseMsg = `Este binário não veio da release oficial do MB CLI (build local, go install, etc.).
O comando mb self update só atualiza binários instalados a partir do GitHub Releases (versão embutida no executável).

Para instalar ou atualizar a versão estável:
  curl -sSL https://raw.githubusercontent.com/carlosdorneles-mb/mb-cli/main/install.sh | bash

Releases: https://github.com/carlosdorneles-mb/mb-cli/releases
`

func newSelfUpdateCmd(deps deps.Dependencies) *cobra.Command {
	var checkOnly bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza o MB CLI para a última release estável do GitHub",
		Long: `Só se aplica a binários da release oficial (versão definida no build). Builds locais recebem uma mensagem a orientar o install.sh.

Consulta o GitHub, compara com a versão embutida e, se houver release mais nova, baixa o binário mb (SHA256) e substitui o executável.

Só atualiza o binário mb (não reinstala gum, glow, jq nem fzf). Linux/macOS, amd64/arm64.

Com --check-only apenas verifica se há release mais nova (sem download). Códigos de saída: 0 = ok; 2 = há atualização; 1 = erro.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			quiet := deps.Runtime != nil && deps.Runtime.Quiet
			if !version.IsReleaseBuild() {
				if !quiet {
					fmt.Fprint(cmd.OutOrStdout(), selfUpdateNonReleaseMsg)
				}
				return nil
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			local := strings.TrimSpace(version.Version)
			if checkOnly {
				out, code, err := selfupdate.RunCheckOnly(ctx, &selfupdate.Config{}, local)
				if out != "" && !quiet {
					fmt.Fprint(cmd.OutOrStdout(), out)
				}
				if err != nil {
					return err
				}
				if code == selfupdate.ExitCodeUpdateAvailable {
					os.Exit(selfupdate.ExitCodeUpdateAvailable)
				}
				return nil
			}
			out, err := selfupdate.Run(ctx, &selfupdate.Config{}, local)
			if out != "" && !quiet {
				fmt.Fprint(cmd.OutOrStdout(), out)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "Só verifica se há atualização (sem baixar)")
	return cmd
}
