package envs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"mb/internal/shared/system"
)

func formatJSON(w io.Writer, rows []ListRow) error {
	obj := make(map[string]string, len(rows))
	for _, r := range rows {
		obj[r.Key] = r.Value
	}
	b, mErr := json.Marshal(obj)
	if mErr != nil {
		return mErr
	}
	_, err := fmt.Fprintln(w, string(b))
	return err
}

func formatText(w io.Writer, rows []ListRow) error {
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%s=%s\n", r.Key, r.Value); err != nil {
			return err
		}
	}
	return nil
}

func formatTable(ctx context.Context, w io.Writer, rows []ListRow) error {
	table := make([][]string, len(rows))
	for i, r := range rows {
		table[i] = []string{r.Key + "=" + r.Value, r.Vault, r.Storage}
	}
	headers := []string{"VAR", "VAULT", "ARMAZENAMENTO"}
	return system.GumTable(ctx, headers, table, w)
}
