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
CREATE INDEX IF NOT EXISTS idx_fichas_acs_id ON fichas (acs_id);
