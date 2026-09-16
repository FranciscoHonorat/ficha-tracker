package repository

import (
	"path/filepath"
	"testing"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userSchema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`

func newTestUserRepository(t *testing.T) *SQLiteUserRepository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath, userSchema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return NewSQLiteUserRepository(db)
}

func TestSQLiteUserRepository_SaveFindCount(t *testing.T) {
	repo := newTestUserRepository(t)

	count, err := repo.CountAll()
	require.NoError(t, err)
	assert.Zero(t, count)

	user, err := domain.NewUser(uuid.New(), "admin", "hashed-password")
	require.NoError(t, err)
	require.NoError(t, repo.Save(user))

	count, err = repo.CountAll()
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	found, err := repo.FindByUsername("Admin")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "hashed-password", found.PasswordHash)

	missing, err := repo.FindByUsername("nobody")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestSQLiteUserRepository_UpdatePasswordHash(t *testing.T) {
	repo := newTestUserRepository(t)

	user, err := domain.NewUser(uuid.New(), "admin", "hashed-password")
	require.NoError(t, err)
	require.NoError(t, repo.Save(user))

	require.NoError(t, repo.UpdatePasswordHash(user.ID, "new-hash"))

	found, err := repo.FindByUsername("admin")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "new-hash", found.PasswordHash)
}
