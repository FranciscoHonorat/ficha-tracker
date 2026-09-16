package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Migrate upgrades a pre-v2.0 database (fichas.acs as free text, no acs/users
// tables) to the current schema. It is idempotent: it is a no-op once
// fichas.acs_id exists, which is already true for a fresh database created
// from schema.sql.
func Migrate(db *sql.DB) error {
	hasACSID, err := hasColumn(db, "fichas", "acs_id")
	if err != nil {
		return fmt.Errorf("checking fichas.acs_id: %w", err)
	}
	if hasACSID {
		return ensureACSIDIndex(db)
	}

	if err := migrateLegacyACS(db); err != nil {
		return err
	}

	return ensureACSIDIndex(db)
}

func ensureACSIDIndex(db *sql.DB) error {
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_fichas_acs_id ON fichas (acs_id)`); err != nil {
		return fmt.Errorf("creating fichas.acs_id index: %w", err)
	}
	return nil
}

func migrateLegacyACS(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning migration: %w", err)
	}
	defer tx.Rollback()

	for _, stmt := range []string{
		`ALTER TABLE fichas ADD COLUMN acs_id TEXT`,
		`ALTER TABLE fichas ADD COLUMN phone TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE fichas ADD COLUMN notified INTEGER NOT NULL DEFAULT 1`,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("running %q: %w", stmt, err)
		}
	}

	rows, err := tx.Query(`SELECT DISTINCT acs FROM fichas WHERE acs IS NOT NULL AND acs != ''`)
	if err != nil {
		return fmt.Errorf("listing legacy acs names: %w", err)
	}
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("scanning legacy acs name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("reading legacy acs names: %w", err)
	}
	rows.Close()

	for _, name := range names {
		var existingID string
		err := tx.QueryRow(`SELECT id FROM acs WHERE name = ? COLLATE NOCASE`, name).Scan(&existingID)
		switch {
		case err == sql.ErrNoRows:
			if _, err := tx.Exec(
				`INSERT INTO acs (id, name, phone, created_at) VALUES (?, ?, '', datetime('now'))`,
				uuid.New().String(), name,
			); err != nil {
				return fmt.Errorf("backfilling acs %q: %w", name, err)
			}
		case err != nil:
			return fmt.Errorf("looking up acs %q: %w", name, err)
		}
	}

	if _, err := tx.Exec(
		`UPDATE fichas SET acs_id = (SELECT id FROM acs WHERE acs.name = fichas.acs COLLATE NOCASE) WHERE acs_id IS NULL`,
	); err != nil {
		return fmt.Errorf("backfilling fichas.acs_id: %w", err)
	}

	return tx.Commit()
}

func hasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			ctype      string
			notNull    int
			dfltValue  sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dfltValue, &primaryKey); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
