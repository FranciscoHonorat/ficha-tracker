package repository

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const legacySchema = `
CREATE TABLE IF NOT EXISTS fichas (
	id TEXT PRIMARY KEY,
	full_name TEXT NOT NULL,
	request_type TEXT NOT NULL,
	acs TEXT NOT NULL,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS acs (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	phone TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`

func TestMigrate_BackfillsLegacyACS(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	db, err := OpenDB(dbPath, legacySchema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	johnID := uuid.New().String()
	janeID := uuid.New().String()
	_, err = db.Exec(
		`INSERT INTO fichas (id, full_name, request_type, acs, created_at) VALUES (?, ?, ?, ?, ?)`,
		johnID, "John Doe", "Exame de sangue", "Maria ACS", "2024-01-15T10:00:00Z",
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO fichas (id, full_name, request_type, acs, created_at) VALUES (?, ?, ?, ?, ?)`,
		janeID, "Jane Smith", "RX", "maria acs", "2024-01-16T10:00:00Z",
	)
	require.NoError(t, err)

	require.NoError(t, Migrate(db))
	// Migrate must be idempotent.
	require.NoError(t, Migrate(db))

	var acsCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM acs`).Scan(&acsCount))
	assert.Equal(t, 1, acsCount, "both fichas share the same ACS name case-insensitively")

	var johnACSID, janeACSID string
	require.NoError(t, db.QueryRow(`SELECT acs_id FROM fichas WHERE id = ?`, johnID).Scan(&johnACSID))
	require.NoError(t, db.QueryRow(`SELECT acs_id FROM fichas WHERE id = ?`, janeID).Scan(&janeACSID))
	assert.Equal(t, johnACSID, janeACSID)
	assert.NotEmpty(t, johnACSID)

	var phone string
	var notified bool
	require.NoError(t, db.QueryRow(`SELECT phone, notified FROM fichas WHERE id = ?`, johnID).Scan(&phone, &notified))
	assert.Equal(t, "", phone)
	assert.True(t, notified)
}

func TestMigrate_NoopOnFreshSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fresh.db")
	// A fresh schema already has acs_id; feeding it through Migrate must be a no-op.
	db, err := OpenDB(dbPath, `
CREATE TABLE fichas (
	id TEXT PRIMARY KEY, full_name TEXT NOT NULL, request_type TEXT NOT NULL,
	acs_id TEXT NOT NULL, phone TEXT NOT NULL DEFAULT '', notified INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);`)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	assert.NoError(t, Migrate(db))
}
