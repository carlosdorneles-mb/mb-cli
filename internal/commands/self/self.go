package self

import (
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/commands/config"
)

func NewSelfCmd(deps config.Dependencies) *cobra.Command {
	selfCmd := &cobra.Command{
		Use:     "self",
		Short:   "Gerencia operações internas do MB CLI",
		GroupID: "commands",
	}
	selfCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	syncCmd := newSelfSyncCmd(deps)
	syncCmd.GroupID = "commands"
	selfCmd.AddCommand(syncCmd)
	envCmd := newSelfEnvCmd(deps)
	envCmd.GroupID = "commands"
	selfCmd.AddCommand(envCmd)
	return selfCmd
}

// FirstPathSegment returns the first segment of path (before the first "/"), or path if no "/".
func FirstPathSegment(path string) string {
	if path == "" {
		return ""
	}
	idx := strings.Index(path, "/")
	if idx == -1 {
		return path
	}
	return path[:idx]
}
