package repository

import (
	"database/sql"
	"fmt"
	"os"
)

// BackupRepository writes consistent point-in-time snapshots of the
// database. It implements service.BackupRepository.
type BackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db: db}
}

// Backup writes a consistent snapshot of the database to destPath, using
// SQLite's VACUUM INTO so it's safe even with the app's own connection open.
// VACUUM INTO refuses to write over an existing file, but the save dialog
// that supplies destPath already lets the user choose to overwrite one, so
// any existing file at destPath is removed first.
func (r *BackupRepository) Backup(destPath string) error {
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing previous backup at destination: %w", err)
	}
	if _, err := r.db.Exec(`VACUUM INTO ?`, destPath); err != nil {
		return fmt.Errorf("backing up database: %w", err)
	}
	return nil
}
