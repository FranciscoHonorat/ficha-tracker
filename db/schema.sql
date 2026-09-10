CREATE TABLE IF NOT EXISTS fichas (
    id TEXT PRIMARY KEY,
    full_name TEXT NOT NULL,
    request_type TEXT NOT NULL,
    acs TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fichas_full_name ON fichas (full_name);
CREATE INDEX IF NOT EXISTS idx_fichas_created_at ON fichas (created_at);
