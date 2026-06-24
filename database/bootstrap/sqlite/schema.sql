CREATE TABLE IF NOT EXISTS bootstrap_schema_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schema_key TEXT NOT NULL UNIQUE,
    schema_version TEXT NOT NULL,
    schema_checksum TEXT NOT NULL,
    app_version TEXT NOT NULL,
    initialized_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS schema_revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schema_key TEXT NOT NULL,
    version TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT 'baseline',
    checksum TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'applied',
    app_version TEXT NOT NULL DEFAULT '',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    error_message TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schema_key, version, direction)
);

CREATE INDEX IF NOT EXISTS idx_schema_revisions_schema_key ON schema_revisions(schema_key);
CREATE INDEX IF NOT EXISTS idx_schema_revisions_status ON schema_revisions(status);

CREATE TABLE IF NOT EXISTS db_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL DEFAULT 'sqlite',
    host TEXT NOT NULL DEFAULT 'localhost',
    port INTEGER NOT NULL DEFAULT 3306,
    user TEXT NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    database TEXT NOT NULL,
    server_port INTEGER NOT NULL DEFAULT 8080,
    encryption_key TEXT NOT NULL DEFAULT '',
    key_length INTEGER NOT NULL DEFAULT 256,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_db_configs_type ON db_configs(type);
