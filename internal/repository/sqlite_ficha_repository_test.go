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
CREATE TABLE IF NOT EXISTS acs (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	phone TEXT NOT NULL,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS fichas (
	id TEXT PRIMARY KEY,
	full_name TEXT NOT NULL,
	request_type TEXT NOT NULL,
	acs_id TEXT NOT NULL,
	phone TEXT NOT NULL DEFAULT '',
	notified INTEGER NOT NULL DEFAULT 1,
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

// seedACS inserts an ACS row directly (bypassing the ACS repository, which
// has its own tests) so ficha tests can satisfy the acs_id foreign key.
func seedACS(t *testing.T, repo *SQLiteFichaRepository, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := repo.db.Exec(
		`INSERT INTO acs (id, name, phone, created_at) VALUES (?, ?, ?, ?)`,
		id.String(), name, "11999999999", time.Now().Format(rfc3339),
	)
	require.NoError(t, err)
	return id
}

func TestSQLiteFichaRepository_SaveAndFind(t *testing.T) {
	repo := newTestRepository(t)
	acsID := seedACS(t, repo, "Maria ACS")

	ficha, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", acsID, "11988887777", true)
	require.NoError(t, err)

	require.NoError(t, repo.Save(ficha))

	fichas, err := repo.Find("", "")
	require.NoError(t, err)
	require.Len(t, fichas, 1)
	assert.True(t, ficha.Equals(fichas[0]))
	assert.Equal(t, "Maria ACS", fichas[0].ACSName)
}

func TestSQLiteFichaRepository_FindByName(t *testing.T) {
	repo := newTestRepository(t)
	acsID := seedACS(t, repo, "Maria ACS")

	john, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", acsID, "", true)
	require.NoError(t, err)
	jane, err := domain.NewFicha(uuid.New(), "Jane Smith", "Encaminhamento", acsID, "", true)
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
	acsID := seedACS(t, repo, "Maria ACS")

	current := &domain.Ficha{
		ID: uuid.New(), FullName: "Current Patient", RequestType: "Consulta",
		ACSID: acsID, Notified: true, CreatedAt: time.Now(),
	}
	past := &domain.Ficha{
		ID: uuid.New(), FullName: "Old Patient", RequestType: "Consulta",
		ACSID: acsID, Notified: true, CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
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
	acsID := seedACS(t, repo, "Maria ACS")

	current := &domain.Ficha{
		ID: uuid.New(), FullName: "Current Patient", RequestType: "Consulta",
		ACSID: acsID, Notified: true, CreatedAt: time.Now(),
	}
	past := &domain.Ficha{
		ID: uuid.New(), FullName: "Old Patient", RequestType: "Consulta",
		ACSID: acsID, Notified: true, CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, repo.Save(current))
	require.NoError(t, repo.Save(past))

	months, err := repo.FindMonths()
	require.NoError(t, err)
	assert.Equal(t, []string{time.Now().Format("2006-01"), "2024-01"}, months)
}

func TestSQLiteFichaRepository_UpdateAndDelete(t *testing.T) {
	repo := newTestRepository(t)
	acsID := seedACS(t, repo, "Maria ACS")

	ficha, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", acsID, "11988887777", false)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ficha))

	ficha.FullName = "John Updated"
	ficha.Notified = true
	require.NoError(t, repo.Update(ficha))

	found, err := repo.Find("", "")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "John Updated", found[0].FullName)
	assert.True(t, found[0].Notified)

	require.NoError(t, repo.Delete(ficha.ID))

	found, err = repo.Find("", "")
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestSQLiteFichaRepository_FindByACSIDAndCount(t *testing.T) {
	repo := newTestRepository(t)
	acsID := seedACS(t, repo, "Maria ACS")
	otherACSID := seedACS(t, repo, "Joana ACS")

	f1, err := domain.NewFicha(uuid.New(), "John Doe", "Exame de sangue", acsID, "", true)
	require.NoError(t, err)
	f2, err := domain.NewFicha(uuid.New(), "Jane Smith", "RX", acsID, "", true)
	require.NoError(t, err)
	f3, err := domain.NewFicha(uuid.New(), "Other Patient", "RX", otherACSID, "", true)
	require.NoError(t, err)
	require.NoError(t, repo.Save(f1))
	require.NoError(t, repo.Save(f2))
	require.NoError(t, repo.Save(f3))

	found, err := repo.FindByACSID(acsID)
	require.NoError(t, err)
	assert.Len(t, found, 2)

	count, err := repo.CountByACSID(acsID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = repo.CountByACSID(otherACSID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
