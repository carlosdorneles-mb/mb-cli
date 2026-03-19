package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"mb/internal/cache"
	"mb/internal/shared/envgroup"
	"mb/internal/shared/safepath"
)

// ValidationWarning represents a plugin that was skipped during scan due to validation errors.
type ValidationWarning struct {
	Path    string // path to manifest.yaml
	Message string // message in Portuguese
}

type Scanner struct {
	pluginsDir string
	// DebugLog receives plugin help inconsistency messages (e.g. invalid group_id). Usually wired to system.Logger.Debug when verbose.
	DebugLog func(msg string)
}

func NewScanner(pluginsDir string) *Scanner {
	return &Scanner{pluginsDir: pluginsDir}
}

// nestedPluginGroupIDRaw returns manifest group_id for nested leaves; top-level returns "" and may log.
func nestedPluginGroupIDRaw(dbCommandPath, manifestGroupID string, debug func(string)) string {
	gid := strings.TrimSpace(manifestGroupID)
	if !strings.Contains(dbCommandPath, "/") {
		if gid != "" && debug != nil {
			debug(
				fmt.Sprintf(
					"plugin help: command_path=%q group_id=%q ignorado (commando top-level fica em PLUGINS)",
					dbCommandPath,
					gid,
				),
			)
		}
		return ""
	}
	return gid
}

// collectHelpGroupBatchesUnderRoot returns one batch per groups.yaml (paths sorted) for global merge at sync.
func collectHelpGroupBatchesUnderRoot(
	rootPath string,
	warnings *[]ValidationWarning,
) [][]HelpGroupDef {
	var paths []string
	_ = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || d.Name() != "groups.yaml" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	sort.Strings(paths)
	var batches [][]HelpGroupDef
	for _, path := range paths {
		defs, err := LoadGroupsFile(path)
		if err != nil {
			*warnings = append(*warnings, ValidationWarning{
				Path:    path,
				Message: "groups.yaml inválido: " + err.Error(),
			})
			continue
		}
		if len(defs) > 0 {
			batches = append(batches, defs)
		}
	}
	return batches
}

// pathSegmentForDir returns a CLI path segment for dir: manifest.command if set, else the directory base name.
func pathSegmentForDir(dir string) (string, error) {
	manifestPath := filepath.Join(dir, "manifest.yaml")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Base(dir), nil
		}
		return "", err
	}
	var m Manifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("%s: %w", manifestPath, err)
	}
	if s := strings.TrimSpace(m.Command); s != "" {
		return s, nil
	}
	return filepath.Base(dir), nil
}

// commandPathForPluginDir builds the slash-separated command path for a plugin or category directory
// relative to rootPath (manifest tree). The last path component is always the leaf directory name.
func commandPathForPluginDir(rootPath, baseDir string) (string, error) {
	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootPath, baseDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("plugin dir %s fora da raiz %s", baseDir, rootPath)
	}
	if rel == "." {
		return "", nil
	}
	parts := strings.Split(rel, string(filepath.Separator))
	var segs []string
	for i := 0; i < len(parts)-1; i++ {
		sub := filepath.Join(append([]string{rootPath}, parts[:i+1]...)...)
		seg, err := pathSegmentForDir(sub)
		if err != nil {
			return "", err
		}
		segs = append(segs, seg)
	}
	segs = append(segs, parts[len(parts)-1])
	inner := strings.Join(segs, "/")
	rootSeg, err := pathSegmentForDir(rootPath)
	if err != nil {
		return "", err
	}
	if rootSeg == "" {
		return inner, nil
	}
	return rootSeg + "/" + inner, nil
}

// categoryPathForDir builds the category path for a directory that has a category-only manifest.
func categoryPathForDir(rootPath, baseDir string) (string, error) {
	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootPath, baseDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("category dir %s fora da raiz %s", baseDir, rootPath)
	}
	if rel == "." {
		return "", nil
	}
	parts := strings.Split(rel, string(filepath.Separator))
	var segs []string
	for i := 0; i < len(parts); i++ {
		sub := filepath.Join(append([]string{rootPath}, parts[:i+1]...)...)
		seg, err := pathSegmentForDir(sub)
		if err != nil {
			return "", err
		}
		segs = append(segs, seg)
	}
	inner := strings.Join(segs, "/")
	rootSeg, err := pathSegmentForDir(rootPath)
	if err != nil {
		return "", err
	}
	if rootSeg == "" {
		return inner, nil
	}
	return rootSeg + "/" + inner, nil
}

// ValidatePluginRoot returns an error if dir/manifest.yaml is missing or invalid.
func ValidatePluginRoot(dir string) error {
	mp := filepath.Join(dir, "manifest.yaml")
	manifest, _, err := LoadManifest(mp)
	if err != nil {
		return err
	}
	if errs := validateManifest(manifest, dir); len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
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
		} else if err := safepath.ValidateUnderDir(execPath, baseDir); err != nil {
			errs = append(errs, "entrypoint fora do diretório do plugin: "+manifest.Entrypoint)
		}
	}
	if manifest.Readme != "" {
		readmePath := filepath.Join(baseDir, manifest.Readme)
		if err := safepath.ValidateUnderDir(readmePath, baseDir); err != nil {
			errs = append(errs, "readme fora do diretório do plugin: "+manifest.Readme)
		}
	}
	for _, e := range manifest.Flags.List {
		if e.Entrypoint != "" {
			flagPath := filepath.Join(baseDir, e.Entrypoint)
			if err := safepath.ValidateUnderDir(flagPath, baseDir); err != nil {
				errs = append(errs, "flag entrypoint fora do diretório do plugin: "+e.Entrypoint)
				break
			}
		}
	}
	if (manifest.Entrypoint != "" || manifest.Flags.Len() > 0) &&
		strings.TrimSpace(manifest.Command) == "" {
		errs = append(errs, "command é obrigatório quando há entrypoint ou flags")
	}
	if manifest.Entrypoint != "" || manifest.Flags.Len() > 0 {
		for _, ef := range manifest.EnvFiles.List {
			if strings.TrimSpace(ef.File) == "" {
				errs = append(errs, "env_files: file não pode ser vazio")
				break
			}
			if err := envgroup.Validate(ef.Group); err != nil {
				errs = append(errs, "env_files grupo inválido ("+ef.File+"): "+err.Error())
				break
			}
			envPath := filepath.Join(baseDir, ef.File)
			if err := safepath.ValidateUnderDir(envPath, baseDir); err != nil {
				errs = append(errs, "env_files fora do diretório do plugin: "+ef.File)
				break
			}
		}
	}
	return errs
}

// cobraFieldsFromManifest returns UseTemplate, ArgsCount, AliasesJSON, Example, LongDescription, Deprecated for cache.Plugin.
func cobraFieldsFromManifest(
	manifest Manifest,
) (useTemplate string, argsCount int, aliasesJSON, example, longDescription, deprecated string, err error) {
	argsCount = manifest.Args
	if argsCount < 0 {
		argsCount = 0
	}
	if len(manifest.Aliases) > 0 {
		b, err := json.Marshal(manifest.Aliases)
		if err != nil {
			return "", 0, "", "", "", "", err
		}
		aliasesJSON = string(b)
	}
	return manifest.Use, argsCount, aliasesJSON, manifest.Example, manifest.LongDescription, manifest.Deprecated, nil
}

func marshalEnvFilesJSON(m Manifest) (string, error) {
	if m.EnvFiles.Len() == 0 {
		return "", nil
	}
	b, err := json.Marshal(m.EnvFiles.List)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Scanner) scanTree(
	rootPath string,
) ([]cache.Plugin, []cache.Category, []ValidationWarning, [][]HelpGroupDef, error) {
	plugins := []cache.Plugin{}
	categories := []cache.Category{}
	warnings := []ValidationWarning{}
	debug := s.DebugLog

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkErr error) error {
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
			warnings = append(
				warnings,
				ValidationWarning{Path: path, Message: strings.Join(errs, "; ")},
			)
			return nil
		}

		commandPath, err := commandPathForPluginDir(rootPath, baseDir)
		if err != nil {
			return err
		}

		commandName := manifest.Command
		if commandName == "" {
			commandName = filepath.Base(baseDir)
		}

		dbCommandPath := commandPath
		if dbCommandPath == "" && (manifest.Entrypoint != "" || manifest.Flags.Len() > 0) {
			dbCommandPath = commandName
		}

		hash := sha256.Sum256(raw)
		configHash := hex.EncodeToString(hash[:])
		readmePath := ""
		if manifest.Readme != "" {
			readmePath = filepath.Join(baseDir, manifest.Readme)
		}

		pluginDir := baseDir

		if manifest.Entrypoint != "" {
			execPath := filepath.Join(baseDir, manifest.Entrypoint)
			pluginType := PluginTypeFromEntrypoint(manifest.Entrypoint)
			flagsJSON := ""
			if manifest.Flags.Len() > 0 {
				flagsMap := manifest.Flags.ToMap()
				flagsJSONBytes, err := json.Marshal(flagsMap)
				if err != nil {
					return fmt.Errorf("marshal flags %s: %w", path, err)
				}
				flagsJSON = string(flagsJSONBytes)
			}
			useT, argsC, aliasesJ, ex, longD, dep := "", 0, "", "", "", ""
			u, a, aj, e, ld, d, err := cobraFieldsFromManifest(manifest)
			if err != nil {
				return fmt.Errorf("cobra fields %s: %w", path, err)
			}
			useT, argsC, aliasesJ, ex, longD, dep = u, a, aj, e, ld, d
			envFilesJ, err := marshalEnvFilesJSON(manifest)
			if err != nil {
				return fmt.Errorf("env_files %s: %w", path, err)
			}
			gid := nestedPluginGroupIDRaw(dbCommandPath, manifest.GroupID, debug)
			plugins = append(plugins, cache.Plugin{
				CommandPath:     dbCommandPath,
				CommandName:     commandName,
				Description:     manifest.Description,
				ExecPath:        execPath,
				PluginType:      pluginType,
				ConfigHash:      configHash,
				ReadmePath:      readmePath,
				FlagsJSON:       flagsJSON,
				UseTemplate:     useT,
				ArgsCount:       argsC,
				AliasesJSON:     aliasesJ,
				Example:         ex,
				LongDescription: longD,
				Deprecated:      dep,
				PluginDir:       pluginDir,
				Hidden:          manifest.Hidden,
				EnvFilesJSON:    envFilesJ,
				GroupID:         gid,
			})
			return nil
		}

		if manifest.Flags.Len() > 0 {
			flagsMap := manifest.Flags.ToMap()
			flagsJSON, err := json.Marshal(flagsMap)
			if err != nil {
				return fmt.Errorf("marshal flags %s: %w", path, err)
			}
			useT, argsC, aliasesJ, ex, longD, dep := "", 0, "", "", "", ""
			u, a, aj, e, ld, d, err := cobraFieldsFromManifest(manifest)
			if err != nil {
				return fmt.Errorf("cobra fields %s: %w", path, err)
			}
			useT, argsC, aliasesJ, ex, longD, dep = u, a, aj, e, ld, d
			envFilesJ, err := marshalEnvFilesJSON(manifest)
			if err != nil {
				return fmt.Errorf("env_files %s: %w", path, err)
			}
			gid := nestedPluginGroupIDRaw(dbCommandPath, manifest.GroupID, debug)
			plugins = append(plugins, cache.Plugin{
				CommandPath:     dbCommandPath,
				CommandName:     commandName,
				Description:     manifest.Description,
				ExecPath:        "",
				PluginType:      "",
				ConfigHash:      configHash,
				ReadmePath:      readmePath,
				FlagsJSON:       string(flagsJSON),
				UseTemplate:     useT,
				ArgsCount:       argsC,
				AliasesJSON:     aliasesJ,
				Example:         ex,
				LongDescription: longD,
				Deprecated:      dep,
				PluginDir:       pluginDir,
				Hidden:          manifest.Hidden,
				EnvFilesJSON:    envFilesJ,
				GroupID:         gid,
			})
			return nil
		}

		catPath, err := categoryPathForDir(rootPath, baseDir)
		if err != nil {
			return err
		}
		if catPath == "" {
			return nil
		}
		catGid := nestedPluginGroupIDRaw(catPath, manifest.GroupID, debug)
		categories = append(categories, cache.Category{
			Path:        catPath,
			Description: manifest.Description,
			ReadmePath:  readmePath,
			Hidden:      manifest.Hidden,
			GroupID:     catGid,
		})
		return nil
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	helpBatches := collectHelpGroupBatchesUnderRoot(rootPath, &warnings)

	rootManifestPath := filepath.Join(rootPath, "manifest.yaml")
	if raw, err := os.ReadFile(rootManifestPath); err == nil {
		var m Manifest
		if err := yaml.Unmarshal(raw, &m); err == nil && m.Entrypoint == "" && m.Flags.Len() == 0 {
			seg, err := pathSegmentForDir(rootPath)
			if err == nil && seg != "" {
				readmePath := ""
				if m.Readme != "" {
					readmePath = filepath.Join(rootPath, m.Readme)
				}
				dup := false
				for _, c := range categories {
					if c.Path == seg {
						dup = true
						break
					}
				}
				if !dup {
					categories = append(categories, cache.Category{
						Path:        seg,
						Description: m.Description,
						ReadmePath:  readmePath,
						Hidden:      m.Hidden,
					})
				}
			}
		}
	}

	return plugins, categories, warnings, helpBatches, nil
}

// Scan percorre cada subpasta imediata de PluginsDir e agrega plugins, categorias e lotes de groups.yaml.
func (s *Scanner) Scan() ([]cache.Plugin, []cache.Category, []ValidationWarning, [][]HelpGroupDef, error) {
	plugins := []cache.Plugin{}
	categories := []cache.Category{}
	warnings := []ValidationWarning{}
	var batches [][]HelpGroupDef
	if _, err := os.Stat(s.pluginsDir); os.IsNotExist(err) {
		return plugins, categories, warnings, batches, nil
	}

	entries, err := os.ReadDir(s.pluginsDir)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		rootPath := filepath.Join(s.pluginsDir, e.Name())
		p, c, w, hg, err := s.scanTree(rootPath)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		plugins = append(plugins, p...)
		categories = append(categories, c...)
		warnings = append(warnings, w...)
		batches = append(batches, hg...)
	}
	return plugins, categories, warnings, batches, nil
}

// ScanDir scans a single directory (local plugin path or one install root). Command paths are
// relative to the manifest tree only; installName is ignored for CLI paths (kept for API compatibility).
func (s *Scanner) ScanDir(
	rootPath string,
	_ string,
) ([]cache.Plugin, []cache.Category, []ValidationWarning, [][]HelpGroupDef, error) {
	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return []cache.Plugin{}, []cache.Category{}, []ValidationWarning{}, nil, nil
	}
	return s.scanTree(rootPath)
}
