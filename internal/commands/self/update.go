package self

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/commands/config"
	"mb/internal/selfupdate"
	"mb/internal/version"
)

func newSelfUpdateCmd(_ config.Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Atualiza o MB CLI para a última release estável do GitHub",
		Long: `Consulta o repositório oficial no GitHub, compara com a versão instalada e,
se existir uma release mais recente, baixa o binário (com verificação SHA256) e substitui o executável atual.

Só atualiza o binário mb (não reinstala gum, glow, jq nem fzf). Suportado em Linux e macOS, amd64 e arm64.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			local := localCLIVersion()
			out, err := selfupdate.Run(ctx, &selfupdate.Config{}, local)
			if out != "" {
				fmt.Fprint(cmd.OutOrStdout(), out)
			}
			return err
		},
	}
}

func localCLIVersion() string {
	if v := strings.TrimSpace(version.Version); v != "" {
		return v
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}
