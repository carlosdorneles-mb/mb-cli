package plugin

// ValidationWarning represents a plugin that was skipped during scan due to validation errors.
type ValidationWarning struct {
	Path    string // path to manifest.yaml
	Message string // message in Portuguese
}
