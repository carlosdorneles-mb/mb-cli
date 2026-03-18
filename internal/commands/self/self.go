package self

import (
	"strings"

	"github.com/spf13/cobra"

	selenv "mb/internal/commands/self/env"
	"mb/internal/deps"
)

func NewSelfCmd(deps deps.Dependencies) *cobra.Command {
	selfCmd := &cobra.Command{
		Use:     "self",
		Aliases: []string{"s"},
		Short:   "Gerencia operações internas do MB CLI",
		GroupID: "commands",
	}
	selfCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	syncCmd := newSelfSyncCmd(deps)
	syncCmd.GroupID = "commands"
	selfCmd.AddCommand(syncCmd)
	envCmd := selenv.NewCmd(deps)
	envCmd.GroupID = "commands"
	selfCmd.AddCommand(envCmd)
	updateCmd := newSelfUpdateCmd(deps)
	updateCmd.GroupID = "commands"
	selfCmd.AddCommand(updateCmd)

	selfCmd.InitDefaultCompletionCmd()
	customizeCompletionPT(selfCmd)

	return selfCmd
}

func customizeCompletionPT(selfCmd *cobra.Command) {
	completionCmd := findCommand(selfCmd.Commands(), "completion")
	if completionCmd == nil {
		return
	}
	completionCmd.Short = "Gera o script de autocompletar do shell"
	completionCmd.GroupID = "commands"
	completionCmd.Long = "Gera o script de autocompletar para o MB CLI para o shell especificado.\nConsulte a ajuda de cada subcomando para detalhes de como usar o script gerado."
	const completionGroupID = "completion_shells"
	completionCmd.AddGroup(&cobra.Group{ID: completionGroupID, Title: "COMANDOS"})
	shortPT := map[string]string{
		"bash":       "Gera o script de autocompletar para bash",
		"zsh":        "Gera o script de autocompletar para zsh",
		"fish":       "Gera o script de autocompletar para fish",
		"powershell": "Gera o script de autocompletar para powershell",
	}
	for _, sub := range completionCmd.Commands() {
		sub.GroupID = completionGroupID
		if short, ok := shortPT[sub.Name()]; ok {
			sub.Short = short
		}
		if f := sub.Flags().Lookup("no-descriptions"); f != nil {
			f.Usage = "Desativa as descrições no autocompletar"
		}
	}
}

func findCommand(cmds []*cobra.Command, name string) *cobra.Command {
	for _, c := range cmds {
		if c.Name() == name {
			return c
		}
	}
	return nil
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
