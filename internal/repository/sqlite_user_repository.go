package repository

import (
	"database/sql"
	"fmt"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

// SQLiteUserRepository persists Users in a SQLite database. It implements
// service.UserRepository.
type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) Save(user *domain.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (id, username, password_hash, created_at) VALUES (?, ?, ?, ?)`,
		user.ID.String(), user.Username, user.PasswordHash, user.CreatedAt.Format(rfc3339),
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}
	return nil
}

// FindByUsername returns the user with the given username, matched
// case-insensitively, or nil if none exists.
func (r *SQLiteUserRepository) FindByUsername(username string) (*domain.User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = ? COLLATE NOCASE`,
		username,
	)

	var (
		idStr, name, passwordHash, createdAtStr string
	)
	err := row.Scan(&idStr, &name, &passwordHash, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by username: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	createdAt, err := parseStoredTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parsing user created_at: %w", err)
	}

	return &domain.User{
		ID:           id,
		Username:     name,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}, nil
}

// CountAll returns how many accounts exist.
func (r *SQLiteUserRepository) CountAll() (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}
