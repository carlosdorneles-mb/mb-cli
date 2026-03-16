package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"mb/internal/cache"
)

// ValidationWarning represents a plugin that was skipped during scan due to validation errors.
type ValidationWarning struct {
	Path    string // path to manifest.yaml
	Message string // message in Portuguese
}

type Scanner struct {
	pluginsDir string
}

func NewScanner(pluginsDir string) *Scanner {
	return &Scanner{pluginsDir: pluginsDir}
}

// validateManifest returns a list of validation errors. Empty slice means valid.
func validateManifest(manifest Manifest, baseDir string) []string {
	var errs []string
	if manifest.Entrypoint != "" {
		execPath := filepath.Join(baseDir, manifest.Entrypoint)
		if _, err := os.Stat(execPath); err != nil {
			if os.IsNotExist(err) {
				errs = append(errs, "entrypoint não encontrado: "+manifest.Entrypoint)
			} else {
				errs = append(errs, "entrypoint inacessível: "+manifest.Entrypoint)
			}
		}
	}
	return errs
}

func (s *Scanner) Scan() ([]cache.Plugin, []cache.Category, []ValidationWarning, error) {
	plugins := []cache.Plugin{}
	categories := []cache.Category{}
	warnings := []ValidationWarning{}
	if _, err := os.Stat(s.pluginsDir); os.IsNotExist(err) {
		return plugins, categories, warnings, nil
	}

	err := filepath.WalkDir(s.pluginsDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "manifest.yaml" {
			return nil
		}

		manifest, raw, err := LoadManifest(path)
		if err != nil {
			return fmt.Errorf("load manifest %s: %w", path, err)
		}

		baseDir := filepath.Dir(path)
		if errs := validateManifest(manifest, baseDir); len(errs) > 0 {
			warnings = append(warnings, ValidationWarning{Path: path, Message: strings.Join(errs, "; ")})
			return nil
		}

		relPath, err := filepath.Rel(s.pluginsDir, baseDir)
		if err != nil {
			return fmt.Errorf("relative path %s: %w", baseDir, err)
		}
		commandPath := filepath.ToSlash(relPath)
		if commandPath == "." {
			commandPath = ""
		}

		commandName := manifest.Command
		if commandName == "" {
			commandName = filepath.Base(baseDir)
		}

		hash := sha256.Sum256(raw)
		configHash := hex.EncodeToString(hash[:])
		readmePath := ""
		if manifest.Readme != "" {
			readmePath = filepath.Join(baseDir, manifest.Readme)
		}

		// Leaf with entrypoint: ignore Flags; type inferred from .sh suffix
		if manifest.Entrypoint != "" {
			execPath := filepath.Join(baseDir, manifest.Entrypoint)
			pluginType := PluginTypeFromEntrypoint(manifest.Entrypoint)
			plugins = append(plugins, cache.Plugin{
				CommandPath: commandPath,
				CommandName: commandName,
				Description: manifest.Description,
				ExecPath:    execPath,
				PluginType:  pluginType,
				ConfigHash:  configHash,
				ReadmePath:  readmePath,
				FlagsJSON:   "",
			})
			return nil
		}

		// Leaf with flags only: no entrypoint, has Flags (list or map)
		if manifest.Flags.Len() > 0 {
			flagsMap := manifest.Flags.ToMap()
			flagsJSON, err := json.Marshal(flagsMap)
			if err != nil {
				return fmt.Errorf("marshal flags %s: %w", path, err)
			}
			plugins = append(plugins, cache.Plugin{
				CommandPath: commandPath,
				CommandName: commandName,
				Description: manifest.Description,
				ExecPath:    "",
				PluginType:  "",
				ConfigHash:  configHash,
				ReadmePath:  readmePath,
				FlagsJSON:   string(flagsJSON),
			})
			return nil
		}

		// Category only (manifest without entrypoint and without flags)
		categories = append(categories, cache.Category{
			Path:        commandPath,
			Description: manifest.Description,
			ReadmePath:  readmePath,
		})
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return plugins, categories, warnings, nil
}

// ScanDir scans a single directory (e.g. a local plugin path) and returns plugins and categories
// with commandPath prefixed by installName. ExecPath and ReadmePath are absolute.
func (s *Scanner) ScanDir(rootPath, installName string) ([]cache.Plugin, []cache.Category, []ValidationWarning, error) {
	plugins := []cache.Plugin{}
	categories := []cache.Category{}
	warnings := []ValidationWarning{}
	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, nil, nil, err
	}
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return plugins, categories, warnings, nil
	}

	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || d.Name() != "manifest.yaml" {
			return nil
		}

		manifest, raw, err := LoadManifest(path)
		if err != nil {
			return fmt.Errorf("load manifest %s: %w", path, err)
		}

		baseDir := filepath.Dir(path)
		if errs := validateManifest(manifest, baseDir); len(errs) > 0 {
			warnings = append(warnings, ValidationWarning{Path: path, Message: strings.Join(errs, "; ")})
			return nil
		}

		relFromRoot, err := filepath.Rel(rootPath, baseDir)
		if err != nil {
			return fmt.Errorf("relative path %s: %w", baseDir, err)
		}
		var commandPath string
		if relFromRoot == "." {
			commandPath = installName
		} else {
			commandPath = filepath.ToSlash(filepath.Join(installName, relFromRoot))
		}

		commandName := manifest.Command
		if commandName == "" {
			commandName = filepath.Base(baseDir)
		}

		hash := sha256.Sum256(raw)
		configHash := hex.EncodeToString(hash[:])
		readmePath := ""
		if manifest.Readme != "" {
			readmePath = filepath.Join(baseDir, manifest.Readme)
		}

		if manifest.Entrypoint != "" {
			execPath := filepath.Join(baseDir, manifest.Entrypoint)
			pluginType := PluginTypeFromEntrypoint(manifest.Entrypoint)
			plugins = append(plugins, cache.Plugin{
				CommandPath: commandPath,
				CommandName: commandName,
				Description: manifest.Description,
				ExecPath:    execPath,
				PluginType:  pluginType,
				ConfigHash:  configHash,
				ReadmePath:  readmePath,
				FlagsJSON:   "",
			})
			return nil
		}

		if manifest.Flags.Len() > 0 {
			flagsMap := manifest.Flags.ToMap()
			flagsJSON, err := json.Marshal(flagsMap)
			if err != nil {
				return fmt.Errorf("marshal flags %s: %w", path, err)
			}
			plugins = append(plugins, cache.Plugin{
				CommandPath: commandPath,
				CommandName: commandName,
				Description: manifest.Description,
				ExecPath:    "",
				PluginType:  "",
				ConfigHash:  configHash,
				ReadmePath:  readmePath,
				FlagsJSON:   string(flagsJSON),
			})
			return nil
		}

		categories = append(categories, cache.Category{
			Path:        commandPath,
			Description: manifest.Description,
			ReadmePath:  readmePath,
		})
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return plugins, categories, warnings, nil
}
