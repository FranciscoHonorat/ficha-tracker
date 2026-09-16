package repository

import (
	"database/sql"
	"fmt"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// OpenDB opens (creating if needed) the SQLite database at path and applies schema.
func OpenDB(path string, schema string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	return db, nil
}

// SQLiteFichaRepository persists Fichas in a SQLite database. It implements
// service.FichaRepository.
type SQLiteFichaRepository struct {
	db *sql.DB
}

func NewSQLiteFichaRepository(db *sql.DB) *SQLiteFichaRepository {
	return &SQLiteFichaRepository{db: db}
}

func (r *SQLiteFichaRepository) Save(ficha *domain.Ficha) error {
	_, err := r.db.Exec(
		`INSERT INTO fichas (id, full_name, request_type, acs_id, phone, notified, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ficha.ID.String(), ficha.FullName, ficha.RequestType, ficha.ACSID.String(), ficha.Phone, ficha.Notified, ficha.CreatedAt.Format(rfc3339),
	)
	if err != nil {
		return fmt.Errorf("saving ficha: %w", err)
	}
	return nil
}

// Update persists changes to an existing ficha.
func (r *SQLiteFichaRepository) Update(ficha *domain.Ficha) error {
	_, err := r.db.Exec(
		`UPDATE fichas SET full_name = ?, request_type = ?, acs_id = ?, phone = ?, notified = ? WHERE id = ?`,
		ficha.FullName, ficha.RequestType, ficha.ACSID.String(), ficha.Phone, ficha.Notified, ficha.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("updating ficha: %w", err)
	}
	return nil
}

// Delete removes a ficha by id.
func (r *SQLiteFichaRepository) Delete(id uuid.UUID) error {
	if _, err := r.db.Exec(`DELETE FROM fichas WHERE id = ?`, id.String()); err != nil {
		return fmt.Errorf("deleting ficha: %w", err)
	}
	return nil
}

// Find returns fichas matching an optional name filter (substring) and an
// optional month filter ("YYYY-MM"), most recent first. An empty string
// skips that filter.
func (r *SQLiteFichaRepository) Find(name, month string) ([]*domain.Ficha, error) {
	query := `SELECT f.id, f.full_name, f.request_type, f.acs_id, f.phone, f.notified, f.created_at, a.name
		FROM fichas f LEFT JOIN acs a ON a.id = f.acs_id WHERE 1=1`
	var args []any

	if name != "" {
		query += ` AND f.full_name LIKE '%' || ? || '%'`
		args = append(args, name)
	}
	if month != "" {
		query += ` AND substr(f.created_at, 1, 7) = ?`
		args = append(args, month)
	}
	query += ` ORDER BY f.created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing fichas: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows, true)
}

// FindByACSID returns every ficha registered for the given ACS, most recent first.
func (r *SQLiteFichaRepository) FindByACSID(acsID uuid.UUID) ([]*domain.Ficha, error) {
	rows, err := r.db.Query(
		`SELECT f.id, f.full_name, f.request_type, f.acs_id, f.phone, f.notified, f.created_at, a.name
		FROM fichas f LEFT JOIN acs a ON a.id = f.acs_id
		WHERE f.acs_id = ? ORDER BY f.created_at DESC`,
		acsID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("listing fichas by acs: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows, true)
}

// CountByACSID returns how many fichas reference the given ACS.
func (r *SQLiteFichaRepository) CountByACSID(acsID uuid.UUID) (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM fichas WHERE acs_id = ?`, acsID.String()).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting fichas by acs: %w", err)
	}
	return count, nil
}

// FindMonths returns every month ("YYYY-MM") that has at least one ficha
// registered, most recent first.
func (r *SQLiteFichaRepository) FindMonths() ([]string, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT substr(created_at, 1, 7) FROM fichas ORDER BY 1 DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing ficha months: %w", err)
	}
	defer rows.Close()

	var months []string
	for rows.Next() {
		var month string
		if err := rows.Scan(&month); err != nil {
			return nil, fmt.Errorf("scanning ficha month: %w", err)
		}
		months = append(months, month)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading ficha months: %w", err)
	}
	return months, nil
}
