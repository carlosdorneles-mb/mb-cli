CREATE TABLE IF NOT EXISTS plugin_sources (
    install_dir TEXT PRIMARY KEY,
    git_url TEXT,
    ref_type TEXT,
    ref TEXT,
    version TEXT,
    local_path TEXT,
    sub_dir TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
