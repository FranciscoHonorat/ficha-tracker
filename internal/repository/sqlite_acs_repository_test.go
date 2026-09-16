package repository

import (
	"path/filepath"
	"testing"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestACSRepository(t *testing.T) *SQLiteACSRepository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenDB(dbPath, schema)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return NewSQLiteACSRepository(db)
}

func TestSQLiteACSRepository_SaveFindUpdateDelete(t *testing.T) {
	repo := newTestACSRepository(t)

	acs, err := domain.NewACS(uuid.New(), "Maria", "11999999999")
	require.NoError(t, err)
	require.NoError(t, repo.Save(acs))

	all, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, "Maria", all[0].Name)

	found, err := repo.FindByID(acs.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Maria", found.Name)

	byName, err := repo.FindByNameExact("maria")
	require.NoError(t, err)
	require.NotNil(t, byName)
	assert.Equal(t, acs.ID, byName.ID)

	missing, err := repo.FindByNameExact("nobody")
	require.NoError(t, err)
	assert.Nil(t, missing)

	acs.Name = "Maria Silva"
	acs.Phone = "11988887777"
	require.NoError(t, repo.Update(acs))

	found, err = repo.FindByID(acs.ID)
	require.NoError(t, err)
	assert.Equal(t, "Maria Silva", found.Name)
	assert.Equal(t, "11988887777", found.Phone)

	require.NoError(t, repo.Delete(acs.ID))
	found, err = repo.FindByID(acs.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
