package envs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"mb/internal/deps"
	"mb/internal/keyring"
	"mb/internal/system"
)

type envListRow struct {
	key, value, group string
}

func collectEnvListRows(
	d deps.Dependencies,
	listGroup string,
	showSecrets bool,
) ([]envListRow, error) {
	var rows []envListRow
	if listGroup != "" {
		if err := deps.ValidateEnvGroup(listGroup); err != nil {
			return nil, err
		}
		p, err := deps.GroupEnvFilePath(d.Runtime.ConfigDir, listGroup)
		if err != nil {
			return nil, err
		}
		r, err := rowsForPath(p, listGroup, showSecrets)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	defRows, err := rowsForPath(d.Runtime.DefaultEnvPath, "default", showSecrets)
	if err != nil {
		return nil, err
	}
	rows = append(rows, defRows...)
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
		// skip .env.*.secrets files
		if strings.HasSuffix(path, secretsSuffix) {
			continue
		}
		r, err := rowsForPath(path, g, showSecrets)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
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

const secretsSuffix = ".secrets"

func rowsForPath(path, group string, showSecrets bool) ([]envListRow, error) {
	values, err := deps.LoadDefaultEnvValues(path)
	if err != nil {
		return nil, err
	}
	secretKeys, err := deps.LoadSecretKeys(path)
	if err != nil {
		return nil, err
	}
	keyringGroup := envGroupForKeyring(group)
	seen := make(map[string]bool)
	var rows []envListRow
	for _, key := range sortedKeys(values) {
		seen[key] = true
		if isSecret(secretKeys, key) {
			val := "***"
			if showSecrets {
				if v, err := keyring.Get(keyringGroup, key); err == nil {
					val = v
				}
			}
			rows = append(rows, envListRow{key: key, value: val, group: group})
		} else {
			rows = append(rows, envListRow{key: key, value: values[key], group: group})
		}
	}
	for _, key := range secretKeys {
		if seen[key] {
			continue
		}
		val := "***"
		if showSecrets {
			if v, err := keyring.Get(keyringGroup, key); err == nil {
				val = v
			}
		}
		rows = append(rows, envListRow{key: key, value: val, group: group})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].key < rows[j].key })
	return rows, nil
}

func isSecret(secretKeys []string, key string) bool {
	for _, sk := range secretKeys {
		if sk == key {
			return true
		}
	}
	return false
}

func newListCmd(d deps.Dependencies) *cobra.Command {
	var listGroup string
	var asJSON, asText, showSecrets bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "Lista variáveis padrão ou de um grupo específico",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := collectEnvListRows(d, listGroup, showSecrets)
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
		BoolVar(&showSecrets, "show-secrets", false, "Mostra o valor real das variáveis guardadas no keyring (por defeito mostram ***)")
	cmd.Flags().
		BoolVarP(&asJSON, "json", "J", false, "Emite variáveis como objeto JSON {\"CHAVE\":\"valor\",...}")
	cmd.Flags().
		BoolVarP(&asText, "text", "T", false, "Emite somente key=value por linha (sem grupo)")
	cmd.MarkFlagsMutuallyExclusive("json", "text")
	cmd.GroupID = "commands"
	return cmd
}
