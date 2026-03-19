CREATE TABLE IF NOT EXISTS plugins (
    id INTEGER PRIMARY KEY,
    command_path TEXT NOT NULL UNIQUE,
    command_name TEXT NOT NULL,
    description TEXT,
    exec_path TEXT,
    plugin_type TEXT CHECK(plugin_type IS NULL OR plugin_type IN ('sh', 'bin')),
    config_hash TEXT NOT NULL,
    readme_path TEXT,
    flags_json TEXT,
    use_template TEXT,
    args_count INTEGER,
    aliases_json TEXT,
    example TEXT,
    long_description TEXT,
    deprecated TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
