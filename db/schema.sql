CREATE TABLE IF NOT EXISTS acs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_acs_name_nocase ON acs (name COLLATE NOCASE);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_nocase ON users (username COLLATE NOCASE);

-- Records only the moment a ficha or ACS was deleted, nothing about what was
-- deleted. Rows older than 90 days are purged at startup (see
-- internal/repository/sqlite_delete_log_repository.go).
CREATE TABLE IF NOT EXISTS delete_log (
    id TEXT PRIMARY KEY,
    deleted_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_delete_log_deleted_at ON delete_log (deleted_at);

CREATE TABLE IF NOT EXISTS fichas (
    id TEXT PRIMARY KEY,
    full_name TEXT NOT NULL,
    request_type TEXT NOT NULL,
    acs TEXT NOT NULL DEFAULT '',
    acs_id TEXT NOT NULL DEFAULT '' REFERENCES acs(id),
    phone TEXT NOT NULL DEFAULT '',
    notified INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fichas_full_name ON fichas (full_name);
CREATE INDEX IF NOT EXISTS idx_fichas_created_at ON fichas (created_at);
-- idx_fichas_acs_id is created by Migrate() (internal/repository/migrate.go),
-- not here: on a pre-v2.0 database this schema is applied against an
-- existing "fichas" table that doesn't have acs_id yet (CREATE TABLE IF NOT
-- EXISTS is a no-op on it), and Migrate() runs right after to add the
-- column before any index on it can be created.
