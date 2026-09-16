package repository

import (
	"database/sql"
	"fmt"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

// SQLiteACSRepository persists ACS in a SQLite database. It implements
// service.ACSRepository.
type SQLiteACSRepository struct {
	db *sql.DB
}

func NewSQLiteACSRepository(db *sql.DB) *SQLiteACSRepository {
	return &SQLiteACSRepository{db: db}
}

func (r *SQLiteACSRepository) Save(acs *domain.ACS) error {
	_, err := r.db.Exec(
		`INSERT INTO acs (id, name, phone, created_at) VALUES (?, ?, ?, ?)`,
		acs.ID.String(), acs.Name, acs.Phone, acs.CreatedAt.Format(rfc3339),
	)
	if err != nil {
		return fmt.Errorf("saving acs: %w", err)
	}
	return nil
}

func (r *SQLiteACSRepository) Update(acs *domain.ACS) error {
	_, err := r.db.Exec(
		`UPDATE acs SET name = ?, phone = ? WHERE id = ?`,
		acs.Name, acs.Phone, acs.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("updating acs: %w", err)
	}
	return nil
}

func (r *SQLiteACSRepository) Delete(id uuid.UUID) error {
	if _, err := r.db.Exec(`DELETE FROM acs WHERE id = ?`, id.String()); err != nil {
		return fmt.Errorf("deleting acs: %w", err)
	}
	return nil
}

func (r *SQLiteACSRepository) FindAll() ([]*domain.ACS, error) {
	rows, err := r.db.Query(`SELECT id, name, phone, created_at FROM acs ORDER BY name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing acs: %w", err)
	}
	defer rows.Close()

	return scanACSList(rows)
}

func (r *SQLiteACSRepository) FindByID(id uuid.UUID) (*domain.ACS, error) {
	rows, err := r.db.Query(`SELECT id, name, phone, created_at FROM acs WHERE id = ?`, id.String())
	if err != nil {
		return nil, fmt.Errorf("finding acs by id: %w", err)
	}
	defer rows.Close()

	list, err := scanACSList(rows)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

// FindByNameExact returns the ACS with the given name, matched
// case-insensitively, or nil if none exists.
func (r *SQLiteACSRepository) FindByNameExact(name string) (*domain.ACS, error) {
	rows, err := r.db.Query(`SELECT id, name, phone, created_at FROM acs WHERE name = ? COLLATE NOCASE`, name)
	if err != nil {
		return nil, fmt.Errorf("finding acs by name: %w", err)
	}
	defer rows.Close()

	list, err := scanACSList(rows)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}
