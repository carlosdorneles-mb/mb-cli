package env

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/system"
)

func newListCmd(d deps.Dependencies) *cobra.Command {
	var listGroup string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis padrão ou de um grupo específico",
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
	cmd.Flags().StringVar(&listGroup, "group", "", "Lista apenas variáveis do grupo informado")
	cmd.GroupID = "commands"
	return cmd
}
