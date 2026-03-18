package self

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/system"
	"mb/internal/ui"
)

func newSelfEnvCmd(d deps.Dependencies) *cobra.Command {
	selfEnvCmd := &cobra.Command{
		Use:   "env",
		Short: "Gerencia variáveis de ambiente padrão",
	}
	selfEnvCmd.AddGroup(&cobra.Group{ID: "commands", Title: "COMANDOS"})

	var listGroup string
	listCmd := &cobra.Command{
		Use:   "list",
		Aliases: []string{"ls", "l"},
		Short: "Lista variáveis padrão ou de um grupo específico",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var rows [][]string
			if listGroup != "" {
				if err := deps.ValidateEnvGroup(listGroup); err != nil {
					return err
				}
				p, err := deps.GroupEnvFilePath(d.Runtime.ConfigDir, listGroup)
				if err != nil {
					return err
				}
				values, err := deps.LoadDefaultEnvValues(p)
				if err != nil {
					return err
				}
				keys := sortedKeys(values)
				for _, key := range keys {
					rows = append(rows, []string{key + "=" + values[key], listGroup})
				}
			} else {
				defVals, err := deps.LoadDefaultEnvValues(d.Runtime.DefaultEnvPath)
				if err != nil {
					return err
				}
				for _, key := range sortedKeys(defVals) {
					rows = append(rows, []string{key + "=" + defVals[key], "default"})
				}
				matches, err := filepath.Glob(filepath.Join(d.Runtime.ConfigDir, ".env.*"))
				if err != nil {
					return err
				}
				sort.Strings(matches)
				for _, path := range matches {
					base := filepath.Base(path)
					if !strings.HasPrefix(base, ".env.") {
						continue
					}
					g := strings.TrimPrefix(base, ".env.")
					if g == "" || deps.ValidateEnvGroup(g) != nil {
						continue
					}
					vals, err := deps.LoadDefaultEnvValues(path)
					if err != nil {
						return err
					}
					for _, key := range sortedKeys(vals) {
						rows = append(rows, []string{key + "=" + vals[key], g})
					}
				}
				sort.Slice(rows, func(i, j int) bool {
					gi, gj := rows[i][1], rows[j][1]
					if gi != gj {
						if gi == "default" {
							return true
						}
						if gj == "default" {
							return false
						}
						return gi < gj
					}
					return rows[i][0] < rows[j][0]
				})
			}
			headers := []string{"VAR", "GRUPO"}
			return system.GumTable(cmd.Context(), headers, rows, cmd.OutOrStdout())
		},
	}
	listCmd.Flags().StringVar(&listGroup, "group", "", "Lista apenas variáveis do grupo imformado")
	listCmd.GroupID = "commands"
	selfEnvCmd.AddCommand(listCmd)

	var setGroup string
	setCmd := &cobra.Command{
		Use:   "set <KEY> <VALUE>",
		Aliases: []string{"s"},
		Short: "Define uma variável padrão ou pra um grupo específico",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			path, err := envTargetPath(d, setGroup)
			if err != nil {
				return err
			}

			values, err := deps.LoadDefaultEnvValues(path)
			if err != nil {
				return err
			}

			values[key] = value
			if err := deps.SaveDefaultEnvValues(path, values); err != nil {
				return err
			}

			msg := fmt.Sprintf("Variável \"%s\" foi salva no grupo padrão", key)
			if setGroup != "" {
				msg = fmt.Sprintf("Variável \"%s\" foi salva no grupo \"%s\"", key, setGroup)
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(msg))
			return nil
		},
	}
	setCmd.Flags().StringVar(&setGroup, "group", "", "Grava a variável no grupo informado ao invés do grupo padrão")
	setCmd.GroupID = "commands"
	selfEnvCmd.AddCommand(setCmd)

	var unsetGroup string
	unsetCmd := &cobra.Command{
		Use:   "unset <KEY>",
		Aliases: []string{"u"},
		Short: "Remove uma variável padrão ou de um grupo específico",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := envTargetPath(d, unsetGroup)
			if err != nil {
				return err
			}

			values, err := deps.LoadDefaultEnvValues(path)
			if err != nil {
				return err
			}

			delete(values, args[0])
			if err := deps.SaveDefaultEnvValues(path, values); err != nil {
				return err
			}
			msg := fmt.Sprintf("Variável \"%s\" foi removida do grupo padrão", args[0])
			if unsetGroup != "" {
				msg = fmt.Sprintf("Variável \"%s\" foi removida do grupo \"%s\"", args[0], unsetGroup)
			}
			fmt.Fprintln(cmd.OutOrStdout(), ui.RenderSuccess(msg))
			return nil
		},
	}
	unsetCmd.Flags().StringVar(&unsetGroup, "group", "", "Remove do arquivo referente ao grupo informado")
	unsetCmd.GroupID = "commands"
	selfEnvCmd.AddCommand(unsetCmd)

	return selfEnvCmd
}

func envTargetPath(d deps.Dependencies, group string) (string, error) {
	if group == "" {
		return d.Runtime.DefaultEnvPath, nil
	}
	if err := deps.ValidateEnvGroup(group); err != nil {
		return "", err
	}
	return deps.GroupEnvFilePath(d.Runtime.ConfigDir, group)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
