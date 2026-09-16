package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupRepository_Backup(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "source.db")
	db, err := OpenDB(dbPath, deleteLogSchema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`INSERT INTO delete_log (id, deleted_at) VALUES ('1', '2024-01-01T00:00:00Z')`)
	require.NoError(t, err)

	backupPath := filepath.Join(t.TempDir(), "backup.db")
	repo := NewBackupRepository(db)
	require.NoError(t, repo.Backup(backupPath))

	backupDB, err := OpenDB(backupPath, "")
	require.NoError(t, err)
	defer backupDB.Close()

	var count int
	require.NoError(t, backupDB.QueryRow(`SELECT COUNT(*) FROM delete_log`).Scan(&count))
	assert.Equal(t, 1, count)
}

// TestBackupRepository_Backup_OverwritesExistingFile reproduces the native
// save dialog's "overwrite this file?" flow: the user picks a filename that
// already exists. SQLite's VACUUM INTO refuses to write over an existing
// file on its own, so Backup must remove it first.
func TestBackupRepository_Backup_OverwritesExistingFile(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "source.db")
	db, err := OpenDB(dbPath, deleteLogSchema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	backupPath := filepath.Join(t.TempDir(), "backup.db")
	require.NoError(t, os.WriteFile(backupPath, []byte("stale backup"), 0o644))

	repo := NewBackupRepository(db)
	require.NoError(t, repo.Backup(backupPath))

	backupDB, err := OpenDB(backupPath, "")
	require.NoError(t, err)
	defer backupDB.Close()

	var count int
	require.NoError(t, backupDB.QueryRow(`SELECT COUNT(*) FROM delete_log`).Scan(&count))
	assert.Zero(t, count)
}
