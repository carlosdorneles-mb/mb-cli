package ui

import (
	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
)

// MBHelpTheme returns a fang ColorSchemeFunc that applies MB CLI colors:
// titles = orange, commands = green, flags = gray, program name (e.g. "mb" in usage) = orange.
func MBHelpTheme() fang.ColorSchemeFunc {
	return func(c lipgloss.LightDarkFunc) fang.ColorScheme {
		base := fang.DefaultColorScheme(c)
		base.Title = lipgloss.Color("#FFA500")   // orange (align with banner)
		base.Program = lipgloss.Color("#FFA500") // orange (program name in usage)
		base.Command = lipgloss.Color("#00A86B") // green
		base.Flag = lipgloss.Color("#656c76")    // gray
		return base
	}
}
