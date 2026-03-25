package deps

import (
	"os"
	"path/filepath"

	"mb/internal/shared/env"
	"mb/internal/shared/ui"
)

// FileValuesOverlay mutates file-backed env values after BuildEnvFileValues (e.g. manifest env_files).
type FileValuesOverlay func(map[string]string) error

// BuildMergedOSEnviron builds the environment passed to plugin scripts and mb run:
// system env, file layers (via BuildEnvFileValues + optional overlay), --env, gum theme, MB_* verbosity, MB_HELPERS_PATH.
func BuildMergedOSEnviron(d Dependencies, overlay FileValuesOverlay) ([]string, error) {
	return BuildMergedOSEnvironWithExtraInline(d, overlay, nil)
}

// BuildMergedOSEnvironWithExtraInline adds an extra inline KEY=VALUE layer before Runtime --env values.
// Effective precedence: system < file layers (+overlay) < extraInline < --env.
func BuildMergedOSEnvironWithExtraInline(
	d Dependencies,
	overlay FileValuesOverlay,
	extraInline []string,
) ([]string, error) {
	extraValues, err := env.ParseInlinePairs(extraInline)
	if err != nil {
		return nil, err
	}
	cliValues, err := env.ParseInlinePairs(d.Runtime.InlineEnvValues)
	if err != nil {
		return nil, err
	}
	fileValues, err := BuildEnvFileValues(d.Runtime)
	if err != nil {
		return nil, err
	}
	if overlay != nil {
		if err := overlay(fileValues); err != nil {
			return nil, err
		}
	}
	for key, value := range extraValues {
		fileValues[key] = value
	}
	merged := env.Merge(os.Environ(), fileValues, cliValues)
	merged = ui.PrependGumThemeDefaults(merged)
	merged = AppendVerbosityEnv(merged, d.Runtime)
	merged = AppendShellHelpersEnv(merged, d.Runtime.ConfigDir)
	return merged, nil
}

// AppendShellHelpersEnv sets MB_HELPERS_PATH for shell helper scripts.
func AppendShellHelpersEnv(merged []string, configDir string) []string {
	path := filepath.Join(configDir, "lib", "shell")
	return append(merged, "MB_HELPERS_PATH="+path)
}

// AppendVerbosityEnv sets MB_QUIET / MB_VERBOSE when root flags request it.
func AppendVerbosityEnv(merged []string, rt *RuntimeConfig) []string {
	if rt == nil {
		return merged
	}
	if rt.Quiet {
		merged = append(merged, "MB_QUIET=1")
	}
	if rt.Verbose {
		merged = append(merged, "MB_VERBOSE=1")
	}
	return merged
}
