package ui

import "mb/internal/env"

// GumThemeDefaults returns the default GUM_* environment variables for the MB CLI
// theme (orange header/titles, green selected/cursor) so plugins using gum inherit the look.
// Keys can be overridden by env.defaults or --env.
func GumThemeDefaults() map[string]string {
	orange := "#FFA500" // MB banner
	green := "#00A86B"  // MB commands
	grey := "#656c76"   // MB flags
	return map[string]string{
		// choose
		"GUM_CHOOSE_HEADER_FOREGROUND":   orange,
		"GUM_CHOOSE_SELECTED_FOREGROUND": green,
		"GUM_CHOOSE_CURSOR_FOREGROUND":   green,
		// input (evita rosa 212 no cursor/prompt)
		"GUM_INPUT_PROMPT_FOREGROUND":      orange,
		"GUM_INPUT_HEADER_FOREGROUND":      orange,
		"GUM_INPUT_CURSOR_FOREGROUND":      grey,
		"GUM_INPUT_PLACEHOLDER_FOREGROUND": "240",
		// confirm (título laranja, opções sem rosa)
		"GUM_CONFIRM_PROMPT_FOREGROUND":     orange,
		"GUM_CONFIRM_SELECTED_FOREGROUND":   green,
		"GUM_CONFIRM_SELECTED_BACKGROUND":   "235",
		"GUM_CONFIRM_UNSELECTED_FOREGROUND": "254",
		"GUM_CONFIRM_UNSELECTED_BACKGROUND": "235",
		// spin (loading laranja)
		"GUM_SPIN_SPINNER_FOREGROUND": orange,
	}
}

// PrependGumThemeDefaults prepends GUM_* theme defaults to merged only for keys
// not already present, so env.defaults and --env override the CLI theme.
func PrependGumThemeDefaults(merged []string) []string {
	existing := env.AsMap(merged)
	var prefix []string
	for key, value := range GumThemeDefaults() {
		if _, ok := existing[key]; !ok {
			prefix = append(prefix, key+"="+value)
		}
	}
	if len(prefix) == 0 {
		return merged
	}
	return append(prefix, merged...)
}
