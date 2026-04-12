package envs

import (
	"context"
	"io"

	"mb/internal/ports"
)

// FormatType specifies how to render ListRows.
type FormatType int

const (
	FormatTable FormatType = iota
	FormatJSON
	FormatText
)

// ListRequest holds parameters for listing environment variables.
type ListRequest struct {
	Vault       string
	ShowSecrets bool
	Format      FormatType
}

// ListService collects environment variable rows for display.
type ListService struct {
	secrets     ports.SecretStore
	onePassword ports.OnePasswordEnv
	paths       Paths
}

// NewListService creates a new ListService.
func NewListService(
	secrets ports.SecretStore,
	onePassword ports.OnePasswordEnv,
	paths Paths,
) *ListService {
	return &ListService{
		secrets:     secrets,
		onePassword: onePassword,
		paths:       paths,
	}
}

// List returns rows for the requested vault(s).
func (s *ListService) List(ctx context.Context, req ListRequest) ([]ListRow, error) {
	_ = ctx // reserved for future async operations
	return CollectListRows(s.secrets, s.onePassword, s.paths, req.Vault, req.ShowSecrets)
}

// FormatRows writes the rows to w in the requested format.
func FormatRows(
	ctx context.Context,
	w io.Writer,
	rows []ListRow,
	format FormatType,
	showSecrets bool,
	configDir string,
) error {
	switch format {
	case FormatJSON:
		return FormatJSONByVault(w, rows, showSecrets, configDir)
	case FormatText:
		return formatTextPlain(w, rows)
	default:
		return formatTable(ctx, w, rows, configDir)
	}
}

// FormatTextPlain exports the text formatter for CLI use.
func FormatTextPlain(w io.Writer, rows []ListRow) error {
	return formatTextPlain(w, rows)
}
