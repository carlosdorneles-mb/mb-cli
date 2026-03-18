package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/system"
)

type envListRow struct {
	key, value, group string
}

func collectEnvListRows(d deps.Dependencies, listGroup string) ([]envListRow, error) {
	var rows []envListRow
	if listGroup != "" {
		if err := deps.ValidateEnvGroup(listGroup); err != nil {
			return nil, err
		}
		p, err := deps.GroupEnvFilePath(d.Runtime.ConfigDir, listGroup)
		if err != nil {
			return nil, err
		}
		values, err := deps.LoadDefaultEnvValues(p)
		if err != nil {
			return nil, err
		}
		for _, key := range sortedKeys(values) {
			rows = append(rows, envListRow{key: key, value: values[key], group: listGroup})
		}
		return rows, nil
	}

	defVals, err := deps.LoadDefaultEnvValues(d.Runtime.DefaultEnvPath)
	if err != nil {
		return nil, err
	}
	for _, key := range sortedKeys(defVals) {
		rows = append(rows, envListRow{key: key, value: defVals[key], group: "default"})
	}
	matches, err := filepath.Glob(filepath.Join(d.Runtime.ConfigDir, ".env.*"))
	if err != nil {
		return nil, err
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
			return nil, err
		}
		for _, key := range sortedKeys(vals) {
			rows = append(rows, envListRow{key: key, value: vals[key], group: g})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		gi, gj := rows[i].group, rows[j].group
		if gi != gj {
			if gi == "default" {
				return true
			}
			if gj == "default" {
				return false
			}
			return gi < gj
		}
		if rows[i].key != rows[j].key {
			return rows[i].key < rows[j].key
		}
		return rows[i].value < rows[j].value
	})
	return rows, nil
}

func newListCmd(d deps.Dependencies) *cobra.Command {
	var listGroup string
	var asJSON, asText bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis padrão ou de um grupo específico",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := collectEnvListRows(d, listGroup)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch {
			case asJSON:
				obj := make(map[string]string, len(rows))
				for _, r := range rows {
					obj[r.key] = r.value
				}
				b, mErr := json.Marshal(obj)
				if mErr != nil {
					return mErr
				}
				_, err = fmt.Fprintln(out, string(b))
				return err
			case asText:
				for _, r := range rows {
					if _, err = fmt.Fprintf(out, "%s=%s\n", r.key, r.value); err != nil {
						return err
					}
				}
				return nil
			default:
				table := make([][]string, len(rows))
				for i, r := range rows {
					table[i] = []string{r.key + "=" + r.value, r.group}
				}
				headers := []string{"VAR", "GRUPO"}
				return system.GumTable(cmd.Context(), headers, table, out)
			}
		},
	}
	cmd.Flags().StringVar(&listGroup, "group", "", "Lista apenas variáveis do grupo informado")
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite variáveis como objeto JSON {\"CHAVE\":\"valor\",...}")
	cmd.Flags().
		BoolVarP(&asText, "text", "T", false, "Emite somente key=value por linha (sem grupo)")
	cmd.MarkFlagsMutuallyExclusive("json", "text")
	cmd.GroupID = "commands"
	return cmd
}
