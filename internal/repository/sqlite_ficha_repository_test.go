package repository

import (
	"path/filepath"
	"testing"
	"time"

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

func TestSQLiteFichaRepository_SaveAndFind(t *testing.T) {
	repo := newTestRepository(t)

	ficha, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", "Maria ACS")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ficha))

	fichas, err := repo.Find("", "")
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

	found, err := repo.Find("John", "")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "John Doe", found[0].FullName)

	notFound, err := repo.Find("Nobody", "")
	require.NoError(t, err)
	assert.Empty(t, notFound)
}

func TestSQLiteFichaRepository_FindByMonth(t *testing.T) {
	repo := newTestRepository(t)

	current := &domain.Ficha{
		ID: uuid.New(), FullName: "Current Patient", RequestType: "Consulta",
		ACS: "Maria ACS", CreatedAt: time.Now(),
	}
	past := &domain.Ficha{
		ID: uuid.New(), FullName: "Old Patient", RequestType: "Consulta",
		ACS: "Maria ACS", CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, repo.Save(current))
	require.NoError(t, repo.Save(past))

	currentMonth := time.Now().Format("2006-01")

	found, err := repo.Find("", currentMonth)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "Current Patient", found[0].FullName)

	found, err = repo.Find("", "2024-01")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "Old Patient", found[0].FullName)
}

func TestSQLiteFichaRepository_FindMonths(t *testing.T) {
	repo := newTestRepository(t)

	current := &domain.Ficha{
		ID: uuid.New(), FullName: "Current Patient", RequestType: "Consulta",
		ACS: "Maria ACS", CreatedAt: time.Now(),
	}
	past := &domain.Ficha{
		ID: uuid.New(), FullName: "Old Patient", RequestType: "Consulta",
		ACS: "Maria ACS", CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, repo.Save(current))
	require.NoError(t, repo.Save(past))

	months, err := repo.FindMonths()
	require.NoError(t, err)
	assert.Equal(t, []string{time.Now().Format("2006-01"), "2024-01"}, months)
}
