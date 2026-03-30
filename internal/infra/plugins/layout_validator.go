package plugins

import "mb/internal/ports"

// LayoutValidator implements ports.PluginLayoutValidator.
type LayoutValidator struct{}

func (LayoutValidator) ValidatePluginRoot(dir string) error {
	return ValidatePluginRoot(dir)
}

var _ ports.PluginLayoutValidator = LayoutValidator{}
