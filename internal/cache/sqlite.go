package cache

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema_table.sql
var schemaTableSQL string

//go:embed schema_index.sql
var schemaIndexSQL string

//go:embed schema_categories.sql
var schemaCategoriesSQL string

//go:embed schema_plugin_sources.sql
var schemaPluginSourcesSQL string

//go:embed schema_plugin_help_groups.sql
var schemaPluginHelpGroupsSQL string

type Category struct {
	Path        string
	Description string
	ReadmePath  string
	Hidden      bool
	GroupID     string // help group for nested categories (e.g. infra/k8s → INFRAESTRUTURA)
}

type Plugin struct {
	ID              int64
	CommandPath     string // e.g. "infra/ci/deploy"
	CommandName     string
	Description     string
	ExecPath        string // empty for flags-only
	PluginType      string // "sh"|"bin" or "" for flags-only
	ConfigHash      string
	ReadmePath      string
	FlagsJSON       string // for flags-only: JSON map of flag name -> {type, entrypoint}
	UseTemplate     string // Cobra Use (optional)
	ArgsCount       int    // Cobra ExactArgs (0 = no validation)
	AliasesJSON     string // JSON array of strings for Cobra Aliases
	Example         string // Cobra Example
	LongDescription string // Cobra Long
	Deprecated      string // Cobra Deprecated message
	PluginDir       string // absolute path to plugin directory (manifest folder); for execution root
	Hidden          bool   // Cobra Hidden: omit from help, still invokable
	EnvFilesJSON    string // manifest env_files as JSON array of {file, group}
	GroupID         string // help group for nested leaves; empty = default COMANDOS
}

// PluginHelpGroup is a Cobra help section for nested plugin commands (from groups.yaml).
type PluginHelpGroup struct {
	GroupID string
	Title   string
}

// PluginSource represents one installation (top-level dir in PluginsDir or a local path) and its git origin/version or local path.
type PluginSource struct {
	InstallDir string
	GitURL     string
	RefType    string // "tag" | "branch"
	Ref        string // e.g. "v1.2.3" or "main"
	Version    string // current state: tag name or short SHA
	LocalPath  string // when set, plugin is local at this path (not in PluginsDir)
	UpdatedAt  string
}

type Store struct {
	db *sql.DB
}

func NewStore(cacheDBPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(cacheDBPath), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", cacheDBPath)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}
	if err := store.InitSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) InitSchema() error {
	// Migrate old schema (category, subcategory, command_name) to command_path if needed
	if err := s.migrateToCommandPath(); err != nil {
		return err
	}
	if _, err := s.db.Exec(schemaTableSQL); err != nil {
		return err
	}
	if _, err := s.db.Exec(schemaIndexSQL); err != nil {
		return err
	}
	if _, err := s.db.Exec(schemaCategoriesSQL); err != nil {
		return err
	}
	if _, err := s.db.Exec(schemaPluginSourcesSQL); err != nil {
		return err
	}
	if _, err := s.db.Exec(schemaPluginHelpGroupsSQL); err != nil {
		return err
	}
	if err := s.migratePluginSourcesLocalPath(); err != nil {
		return err
	}
	return s.migrateCobraPluginFields()
}

// migrateCobraPluginFields adds use_template, args_count, aliases_json, example, long_description, deprecated to plugins if missing.
func (s *Store) migrateCobraPluginFields() error {
	columns := []struct{ name, typ string }{
		{"use_template", "TEXT"},
		{"args_count", "INTEGER"},
		{"aliases_json", "TEXT"},
		{"example", "TEXT"},
		{"long_description", "TEXT"},
		{"deprecated", "TEXT"},
	}
	for _, c := range columns {
		var has int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name=?", c.name).Scan(&has); err != nil {
			return err
		}
		if has > 0 {
			continue
		}
		if _, err := s.db.Exec("ALTER TABLE plugins ADD COLUMN " + c.name + " " + c.typ); err != nil {
			return err
		}
	}
	if err := s.migratePluginDir(); err != nil {
		return err
	}
	if err := s.migrateHiddenColumns(); err != nil {
		return err
	}
	if err := s.migrateEnvFilesJSONColumn(); err != nil {
		return err
	}
	if err := s.migratePluginGroupIDColumn(); err != nil {
		return err
	}
	return s.migrateCategoryGroupIDColumn()
}

func (s *Store) migrateCategoryGroupIDColumn() error {
	var has int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('categories') WHERE name='group_id'").Scan(&has); err != nil {
		return err
	}
	if has > 0 {
		return nil
	}
	_, err := s.db.Exec("ALTER TABLE categories ADD COLUMN group_id TEXT")
	return err
}

func (s *Store) migratePluginGroupIDColumn() error {
	var has int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name='group_id'").Scan(&has); err != nil {
		return err
	}
	if has > 0 {
		return nil
	}
	_, err := s.db.Exec("ALTER TABLE plugins ADD COLUMN group_id TEXT")
	return err
}

func (s *Store) migrateEnvFilesJSONColumn() error {
	var has int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name='env_files_json'").Scan(&has); err != nil {
		return err
	}
	if has > 0 {
		return nil
	}
	_, err := s.db.Exec("ALTER TABLE plugins ADD COLUMN env_files_json TEXT")
	return err
}

func (s *Store) migratePluginDir() error {
	var has int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name='plugin_dir'").Scan(&has); err != nil {
		return err
	}
	if has > 0 {
		return nil
	}
	_, err := s.db.Exec("ALTER TABLE plugins ADD COLUMN plugin_dir TEXT")
	return err
}

func (s *Store) migrateHiddenColumns() error {
	var has int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name='hidden'").Scan(&has); err != nil {
		return err
	}
	if has == 0 {
		if _, err := s.db.Exec("ALTER TABLE plugins ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0"); err != nil {
			return err
		}
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('categories') WHERE name='hidden'").Scan(&has); err != nil {
		return err
	}
	if has == 0 {
		if _, err := s.db.Exec("ALTER TABLE categories ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) migratePluginSourcesLocalPath() error {
	var has int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugin_sources') WHERE name='local_path'").Scan(&has)
	if has > 0 {
		return nil
	}
	_, err := s.db.Exec("ALTER TABLE plugin_sources ADD COLUMN local_path TEXT")
	return err
}

// migrateToCommandPath converts old plugins table to new schema (command_path, flags_json, etc).
func (s *Store) migrateToCommandPath() error {
	var hasCategory int
	var hasCommandPath int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name='category'").Scan(&hasCategory)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('plugins') WHERE name='command_path'").Scan(&hasCommandPath)
	if hasCategory == 0 || hasCommandPath > 0 {
		return nil
	}
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS plugins_new (
    id INTEGER PRIMARY KEY,
    command_path TEXT NOT NULL UNIQUE,
    command_name TEXT NOT NULL,
    description TEXT,
    exec_path TEXT,
    plugin_type TEXT,
    config_hash TEXT NOT NULL,
    readme_path TEXT,
    flags_json TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO plugins_new (id, command_path, command_name, description, exec_path, plugin_type, config_hash, readme_path, flags_json, updated_at)
SELECT id,
       (COALESCE(category,'') || CASE WHEN subcategory IS NOT NULL AND subcategory != '' THEN '/' || subcategory ELSE '' END || '/' || COALESCE(command_name,'')) AS command_path,
       command_name,
       '' AS description,
       exec_path,
       plugin_type,
       config_hash,
       readme_path,
       NULL AS flags_json,
       COALESCE(updated_at, CURRENT_TIMESTAMP)
FROM plugins;
DROP TABLE plugins;
ALTER TABLE plugins_new RENAME TO plugins;
`)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_plugins_command_path ON plugins (command_path)")
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) UpsertPlugin(plugin Plugin) error {
	// Entrypoint plugin: type sh/bin and exec_path set. Flags-only: flags_json set, exec_path/type can be empty.
	if plugin.FlagsJSON == "" {
		if plugin.PluginType != "sh" && plugin.PluginType != "bin" {
			return fmt.Errorf("invalid plugin type %q", plugin.PluginType)
		}
		if plugin.ExecPath == "" {
			return fmt.Errorf("exec_path required for entrypoint plugin")
		}
	}

	hidden := 0
	if plugin.Hidden {
		hidden = 1
	}
	_, err := s.db.Exec(`
INSERT INTO plugins (command_path, command_name, description, exec_path, plugin_type, config_hash, readme_path, flags_json, use_template, args_count, aliases_json, example, long_description, deprecated, plugin_dir, hidden, env_files_json, group_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(command_path) DO UPDATE SET
  command_name = excluded.command_name,
  description = excluded.description,
  exec_path = excluded.exec_path,
  plugin_type = excluded.plugin_type,
  config_hash = excluded.config_hash,
  readme_path = excluded.readme_path,
  flags_json = excluded.flags_json,
  use_template = excluded.use_template,
  args_count = excluded.args_count,
  aliases_json = excluded.aliases_json,
  example = excluded.example,
  long_description = excluded.long_description,
  deprecated = excluded.deprecated,
  plugin_dir = excluded.plugin_dir,
  hidden = excluded.hidden,
  env_files_json = excluded.env_files_json,
  group_id = excluded.group_id,
  updated_at = CURRENT_TIMESTAMP
`, plugin.CommandPath, plugin.CommandName, plugin.Description, nullEmpty(plugin.ExecPath), nullEmpty(plugin.PluginType), plugin.ConfigHash, nullEmpty(plugin.ReadmePath), nullEmpty(plugin.FlagsJSON),
		nullEmpty(plugin.UseTemplate), plugin.ArgsCount, nullEmpty(plugin.AliasesJSON), nullEmpty(plugin.Example), nullEmpty(plugin.LongDescription), nullEmpty(plugin.Deprecated), nullEmpty(plugin.PluginDir), hidden, nullEmpty(plugin.EnvFilesJSON), nullEmpty(plugin.GroupID))
	return err
}

func nullEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (s *Store) ListPlugins() ([]Plugin, error) {
	rows, err := s.db.Query(`
SELECT id, command_path, command_name, COALESCE(description, ''), COALESCE(exec_path, ''), COALESCE(plugin_type, ''), config_hash, COALESCE(readme_path, ''), COALESCE(flags_json, ''),
  COALESCE(use_template, ''), COALESCE(args_count, 0), COALESCE(aliases_json, ''), COALESCE(example, ''), COALESCE(long_description, ''), COALESCE(deprecated, ''), COALESCE(plugin_dir, ''), COALESCE(hidden, 0), COALESCE(env_files_json, ''), COALESCE(group_id, '')
FROM plugins
ORDER BY command_path
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlugins(rows)
}

func (s *Store) ListByPathPrefix(prefix string) ([]Plugin, error) {
	pattern := prefix + "%"
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		pattern = prefix + "/%"
	}
	rows, err := s.db.Query(`
SELECT id, command_path, command_name, COALESCE(description, ''), COALESCE(exec_path, ''), COALESCE(plugin_type, ''), config_hash, COALESCE(readme_path, ''), COALESCE(flags_json, ''),
  COALESCE(use_template, ''), COALESCE(args_count, 0), COALESCE(aliases_json, ''), COALESCE(example, ''), COALESCE(long_description, ''), COALESCE(deprecated, ''), COALESCE(plugin_dir, ''), COALESCE(hidden, 0), COALESCE(env_files_json, ''), COALESCE(group_id, '')
FROM plugins
WHERE command_path LIKE ? OR command_path = ?
ORDER BY command_path
`, pattern, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPlugins(rows)
}

func (s *Store) UpsertCategory(cat Category) error {
	h := 0
	if cat.Hidden {
		h = 1
	}
	_, err := s.db.Exec(`
INSERT INTO categories (path, description, readme_path, hidden, group_id)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
  description = excluded.description,
  readme_path = excluded.readme_path,
  hidden = excluded.hidden,
  group_id = excluded.group_id
`, cat.Path, nullEmpty(cat.Description), nullEmpty(cat.ReadmePath), h, nullEmpty(cat.GroupID))
	return err
}

func (s *Store) DeleteAllCategories() error {
	_, err := s.db.Exec("DELETE FROM categories")
	return err
}

func (s *Store) DeleteAllPlugins() error {
	_, err := s.db.Exec("DELETE FROM plugins")
	return err
}

func (s *Store) ListCategories() ([]Category, error) {
	rows, err := s.db.Query(`
SELECT path, COALESCE(description, ''), COALESCE(readme_path, ''), COALESCE(hidden, 0), COALESCE(group_id, '')
FROM categories
ORDER BY path
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Category
	for rows.Next() {
		var c Category
		var hid int
		if err := rows.Scan(&c.Path, &c.Description, &c.ReadmePath, &hid, &c.GroupID); err != nil {
			return nil, err
		}
		c.Hidden = hid != 0
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *Store) UpsertPluginSource(ps PluginSource) error {
	_, err := s.db.Exec(`
INSERT INTO plugin_sources (install_dir, git_url, ref_type, ref, version, local_path)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(install_dir) DO UPDATE SET
  git_url = excluded.git_url,
  ref_type = excluded.ref_type,
  ref = excluded.ref,
  version = excluded.version,
  local_path = excluded.local_path,
  updated_at = CURRENT_TIMESTAMP
`, ps.InstallDir, nullEmpty(ps.GitURL), nullEmpty(ps.RefType), nullEmpty(ps.Ref), nullEmpty(ps.Version), nullEmpty(ps.LocalPath))
	return err
}

func (s *Store) GetPluginSource(installDir string) (*PluginSource, error) {
	var ps PluginSource
	err := s.db.QueryRow(`
SELECT install_dir, COALESCE(git_url, ''), COALESCE(ref_type, ''), COALESCE(ref, ''), COALESCE(version, ''), COALESCE(local_path, ''), COALESCE(updated_at, '')
FROM plugin_sources WHERE install_dir = ?
`, installDir).Scan(&ps.InstallDir, &ps.GitURL, &ps.RefType, &ps.Ref, &ps.Version, &ps.LocalPath, &ps.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ps, nil
}

func (s *Store) ListPluginSources() ([]PluginSource, error) {
	rows, err := s.db.Query(`
SELECT install_dir, COALESCE(git_url, ''), COALESCE(ref_type, ''), COALESCE(ref, ''), COALESCE(version, ''), COALESCE(local_path, ''), COALESCE(updated_at, '')
FROM plugin_sources
ORDER BY install_dir
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PluginSource
	for rows.Next() {
		var ps PluginSource
		if err := rows.Scan(&ps.InstallDir, &ps.GitURL, &ps.RefType, &ps.Ref, &ps.Version, &ps.LocalPath, &ps.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, ps)
	}
	return list, rows.Err()
}

func (s *Store) DeletePluginSource(installDir string) error {
	_, err := s.db.Exec("DELETE FROM plugin_sources WHERE install_dir = ?", installDir)
	return err
}

func scanPlugins(rows *sql.Rows) ([]Plugin, error) {
	plugins := []Plugin{}
	for rows.Next() {
		var plugin Plugin
		var hidden int
		if err := rows.Scan(
			&plugin.ID,
			&plugin.CommandPath,
			&plugin.CommandName,
			&plugin.Description,
			&plugin.ExecPath,
			&plugin.PluginType,
			&plugin.ConfigHash,
			&plugin.ReadmePath,
			&plugin.FlagsJSON,
			&plugin.UseTemplate,
			&plugin.ArgsCount,
			&plugin.AliasesJSON,
			&plugin.Example,
			&plugin.LongDescription,
			&plugin.Deprecated,
			&plugin.PluginDir,
			&hidden,
			&plugin.EnvFilesJSON,
			&plugin.GroupID,
		); err != nil {
			return nil, err
		}
		plugin.Hidden = hidden != 0
		plugins = append(plugins, plugin)
	}
	return plugins, rows.Err()
}

func (s *Store) DeleteAllPluginHelpGroups() error {
	_, err := s.db.Exec("DELETE FROM plugin_help_groups")
	return err
}

func (s *Store) UpsertPluginHelpGroup(g PluginHelpGroup) error {
	_, err := s.db.Exec(`
INSERT INTO plugin_help_groups (group_id, title) VALUES (?, ?)
ON CONFLICT(group_id) DO UPDATE SET title = excluded.title
`, g.GroupID, g.Title)
	return err
}

func (s *Store) ListPluginHelpGroups() ([]PluginHelpGroup, error) {
	rows, err := s.db.Query(`SELECT group_id, title FROM plugin_help_groups ORDER BY group_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []PluginHelpGroup
	for rows.Next() {
		var g PluginHelpGroup
		if err := rows.Scan(&g.GroupID, &g.Title); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}
