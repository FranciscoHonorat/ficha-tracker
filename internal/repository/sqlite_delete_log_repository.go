package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeletionRetention is how long a delete_log entry is kept before PurgeOld
// removes it.
const DeletionRetention = 90 * 24 * time.Hour

// SQLiteDeleteLogRepository records the moment a ficha or ACS was deleted,
// nothing else, for a bounded-retention audit trail. It implements
// service.DeleteLogRepository.
type SQLiteDeleteLogRepository struct {
	db *sql.DB
}

func NewSQLiteDeleteLogRepository(db *sql.DB) *SQLiteDeleteLogRepository {
	return &SQLiteDeleteLogRepository{db: db}
}

// Record logs that a deletion happened now.
func (r *SQLiteDeleteLogRepository) Record() error {
	_, err := r.db.Exec(
		`INSERT INTO delete_log (id, deleted_at) VALUES (?, ?)`,
		uuid.New().String(), time.Now().Format(rfc3339),
	)
	if err != nil {
		return fmt.Errorf("recording deletion: %w", err)
	}
	return nil
}

// PurgeOld removes delete_log entries older than DeletionRetention.
func (r *SQLiteDeleteLogRepository) PurgeOld() error {
	cutoff := time.Now().Add(-DeletionRetention).Format(rfc3339)
	if _, err := r.db.Exec(`DELETE FROM delete_log WHERE deleted_at < ?`, cutoff); err != nil {
		return fmt.Errorf("purging old deletion records: %w", err)
	}
	return nil
}

// CountAll returns how many deletion records are currently retained.
func (r *SQLiteDeleteLogRepository) CountAll() (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM delete_log`).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting deletion records: %w", err)
	}
	return count, nil
}
