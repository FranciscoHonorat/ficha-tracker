package repository

import (
	"database/sql"
	"fmt"

	"ficha-tracker/internal/domain"

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
		`INSERT INTO fichas (id, full_name, request_type, acs, created_at) VALUES (?, ?, ?, ?, ?)`,
		ficha.ID.String(), ficha.FullName, ficha.RequestType, ficha.ACS, ficha.CreatedAt.Format(rfc3339),
	)
	if err != nil {
		return fmt.Errorf("saving ficha: %w", err)
	}
	return nil
}

// Find returns fichas matching an optional name filter (substring) and an
// optional month filter ("YYYY-MM"), most recent first. An empty string
// skips that filter.
func (r *SQLiteFichaRepository) Find(name, month string) ([]*domain.Ficha, error) {
	query := `SELECT id, full_name, request_type, acs, created_at FROM fichas WHERE 1=1`
	var args []interface{}

	if name != "" {
		query += ` AND full_name LIKE '%' || ? || '%'`
		args = append(args, name)
	}
	if month != "" {
		query += ` AND substr(created_at, 1, 7) = ?`
		args = append(args, month)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing fichas: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows)
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

func scanFichas(rows *sql.Rows) ([]*domain.Ficha, error) {
	var fichas []*domain.Ficha
	for rows.Next() {
		ficha, err := scanFicha(rows)
		if err != nil {
			return nil, err
		}
		fichas = append(fichas, ficha)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading fichas: %w", err)
	}
	return fichas, nil
}
