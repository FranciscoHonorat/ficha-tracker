package repository

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const deleteLogSchema = `
CREATE TABLE IF NOT EXISTS delete_log (
	id TEXT PRIMARY KEY,
	deleted_at TEXT NOT NULL
);
`

func newTestDeleteLogRepository(t *testing.T) *SQLiteDeleteLogRepository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath, deleteLogSchema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return NewSQLiteDeleteLogRepository(db)
}

func TestSQLiteDeleteLogRepository_RecordAndCount(t *testing.T) {
	repo := newTestDeleteLogRepository(t)

	count, err := repo.CountAll()
	require.NoError(t, err)
	assert.Zero(t, count)

	require.NoError(t, repo.Record())
	require.NoError(t, repo.Record())

	count, err = repo.CountAll()
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSQLiteDeleteLogRepository_PurgeOld(t *testing.T) {
	repo := newTestDeleteLogRepository(t)

	old := time.Now().Add(-100 * 24 * time.Hour).Format(rfc3339)
	recent := time.Now().Add(-10 * 24 * time.Hour).Format(rfc3339)
	_, err := repo.db.Exec(`INSERT INTO delete_log (id, deleted_at) VALUES (?, ?)`, uuid.New().String(), old)
	require.NoError(t, err)
	_, err = repo.db.Exec(`INSERT INTO delete_log (id, deleted_at) VALUES (?, ?)`, uuid.New().String(), recent)
	require.NoError(t, err)

	require.NoError(t, repo.PurgeOld())

	count, err := repo.CountAll()
	require.NoError(t, err)
	assert.Equal(t, 1, count, "only the record older than 90 days should be purged")
}
