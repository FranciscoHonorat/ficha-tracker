package service

import (
	"testing"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeACSRepository struct {
	saved []*domain.ACS
}

func (r *fakeACSRepository) Save(acs *domain.ACS) error {
	r.saved = append(r.saved, acs)
	return nil
}

func (r *fakeACSRepository) Update(acs *domain.ACS) error {
	for i, a := range r.saved {
		if a.ID == acs.ID {
			r.saved[i] = acs
			return nil
		}
	}
	return nil
}

func (r *fakeACSRepository) Delete(id uuid.UUID) error {
	for i, a := range r.saved {
		if a.ID == id {
			r.saved = append(r.saved[:i], r.saved[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *fakeACSRepository) FindAll() ([]*domain.ACS, error) {
	return r.saved, nil
}

func (r *fakeACSRepository) FindByID(id uuid.UUID) (*domain.ACS, error) {
	for _, a := range r.saved {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (r *fakeACSRepository) FindByNameExact(name string) (*domain.ACS, error) {
	for _, a := range r.saved {
		if equalFold(a.Name, name) {
			return a, nil
		}
	}
	return nil, nil
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ac, bc := a[i], b[i]
		if 'A' <= ac && ac <= 'Z' {
			ac += 'a' - 'A'
		}
		if 'A' <= bc && bc <= 'Z' {
			bc += 'a' - 'A'
		}
		if ac != bc {
			return false
		}
	}
	return true
}

type fakeFichaCounter struct {
	counts map[uuid.UUID]int
}

func (c *fakeFichaCounter) CountByACSID(acsID uuid.UUID) (int, error) {
	return c.counts[acsID], nil
}

func newTestACSService() (*ACSService, *fakeACSRepository, *fakeFichaCounter) {
	repo := &fakeACSRepository{}
	counter := &fakeFichaCounter{counts: map[uuid.UUID]int{}}
	return NewACSService(repo, counter, &fakeDeleteLogRepository{}), repo, counter
}

func TestACSService_RegisterACS(t *testing.T) {
	t.Run("registers a valid ACS", func(t *testing.T) {
		service, repo, _ := newTestACSService()

		acs, err := service.RegisterACS("Maria", "11999999999")

		assert.NoError(t, err)
		assert.NotNil(t, acs)
		assert.Len(t, repo.saved, 1)
	})

	t.Run("rejects a duplicate name case-insensitively", func(t *testing.T) {
		service, _, _ := newTestACSService()

		_, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)

		_, err = service.RegisterACS("maria", "11988887777")
		assert.ErrorIs(t, err, domain.ErrDuplicateACSName)
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		service, _, _ := newTestACSService()

		_, err := service.RegisterACS("", "11999999999")
		assert.ErrorIs(t, err, domain.ErrInvalidACSName)
	})
}

func TestACSService_UpdateACS(t *testing.T) {
	t.Run("updates a valid ACS", func(t *testing.T) {
		service, _, _ := newTestACSService()

		acs, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)

		updated, err := service.UpdateACS(acs.ID, "Maria Silva", "11988887777")
		require.NoError(t, err)
		assert.Equal(t, "Maria Silva", updated.Name)
	})

	t.Run("allows keeping the same name on update", func(t *testing.T) {
		service, _, _ := newTestACSService()

		acs, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)

		_, err = service.UpdateACS(acs.ID, "Maria", "11988887777")
		assert.NoError(t, err)
	})

	t.Run("rejects renaming to another ACS's name", func(t *testing.T) {
		service, _, _ := newTestACSService()

		_, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)
		joana, err := service.RegisterACS("Joana", "11988887777")
		require.NoError(t, err)

		_, err = service.UpdateACS(joana.ID, "Maria", "11988887777")
		assert.ErrorIs(t, err, domain.ErrDuplicateACSName)
	})
}

func TestACSService_DeleteACS(t *testing.T) {
	t.Run("deletes an ACS with no fichas", func(t *testing.T) {
		service, repo, _ := newTestACSService()

		acs, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)

		require.NoError(t, service.DeleteACS(acs.ID))
		assert.Empty(t, repo.saved)
	})

	t.Run("refuses to delete an ACS with fichas", func(t *testing.T) {
		service, _, counter := newTestACSService()

		acs, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)
		counter.counts[acs.ID] = 2

		err = service.DeleteACS(acs.ID)
		assert.ErrorIs(t, err, domain.ErrACSHasFichas)
	})

	t.Run("logs the deletion", func(t *testing.T) {
		repo := &fakeACSRepository{}
		counter := &fakeFichaCounter{counts: map[uuid.UUID]int{}}
		deleteLog := &fakeDeleteLogRepository{}
		service := NewACSService(repo, counter, deleteLog)

		acs, err := service.RegisterACS("Maria", "11999999999")
		require.NoError(t, err)

		require.NoError(t, service.DeleteACS(acs.ID))
		assert.Equal(t, 1, deleteLog.records)
	})
}

func TestACSService_ListACS(t *testing.T) {
	service, _, _ := newTestACSService()

	_, _ = service.RegisterACS("Maria", "11999999999")
	_, _ = service.RegisterACS("Joana", "11988887777")

	list, err := service.ListACS()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}
