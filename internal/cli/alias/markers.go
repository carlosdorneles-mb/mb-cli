package alias

// Markers delimit the user-aliases block in shell profile files (idempotent install).
const (
	BlockBegin = "# mb-cli user aliases BEGIN"
	BlockEnd   = "# mb-cli user aliases END"
)
