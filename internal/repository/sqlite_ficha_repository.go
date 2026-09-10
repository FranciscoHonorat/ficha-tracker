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

func (r *SQLiteFichaRepository) FindAll() ([]*domain.Ficha, error) {
	rows, err := r.db.Query(
		`SELECT id, full_name, request_type, acs, created_at FROM fichas ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing fichas: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows)
}

func (r *SQLiteFichaRepository) FindByName(name string) ([]*domain.Ficha, error) {
	rows, err := r.db.Query(
		`SELECT id, full_name, request_type, acs, created_at FROM fichas WHERE full_name LIKE '%' || ? || '%' ORDER BY created_at DESC`,
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("searching fichas: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows)
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
