package repository

import (
	"path/filepath"
	"testing"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const schema = `
CREATE TABLE IF NOT EXISTS fichas (
	id TEXT PRIMARY KEY,
	full_name TEXT NOT NULL,
	request_type TEXT NOT NULL,
	acs TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`

func newTestRepository(t *testing.T) *SQLiteFichaRepository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath, schema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return NewSQLiteFichaRepository(db)
}

func TestSQLiteFichaRepository_SaveAndFindAll(t *testing.T) {
	repo := newTestRepository(t)

	ficha, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", "Maria ACS")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ficha))

	fichas, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, fichas, 1)
	assert.True(t, ficha.Equals(fichas[0]))
}

func TestSQLiteFichaRepository_FindByName(t *testing.T) {
	repo := newTestRepository(t)

	john, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", "Maria ACS")
	require.NoError(t, err)
	jane, err := domain.NewFicha(uuid.New(), "Jane Smith", "Encaminhamento", "Maria ACS")
	require.NoError(t, err)
	require.NoError(t, repo.Save(john))
	require.NoError(t, repo.Save(jane))

	found, err := repo.FindByName("John")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "John Doe", found[0].FullName)

	notFound, err := repo.FindByName("Nobody")
	require.NoError(t, err)
	assert.Empty(t, notFound)
}
