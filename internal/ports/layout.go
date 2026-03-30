package ports

// PluginLayoutValidator checks that a directory is a valid plugin root (manifest tree).
type PluginLayoutValidator interface {
	ValidatePluginRoot(dir string) error
}
