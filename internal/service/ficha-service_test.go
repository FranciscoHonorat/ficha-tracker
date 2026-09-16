package service

import (
	"errors"
	"testing"
	"time"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFichaRepository struct {
	saved   []*domain.Ficha
	saveErr error
	findErr error
}

func (r *fakeFichaRepository) Save(ficha *domain.Ficha) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = append(r.saved, ficha)
	return nil
}

func (r *fakeFichaRepository) Update(ficha *domain.Ficha) error {
	for i, f := range r.saved {
		if f.ID == ficha.ID {
			r.saved[i] = ficha
			return nil
		}
	}
	return errors.New("not found")
}

func (r *fakeFichaRepository) Delete(id uuid.UUID) error {
	for i, f := range r.saved {
		if f.ID == id {
			r.saved = append(r.saved[:i], r.saved[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *fakeFichaRepository) Find(name, month string) ([]*domain.Ficha, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	var result []*domain.Ficha
	for _, f := range r.saved {
		if name != "" && f.FullName != name {
			continue
		}
		if month != "" && f.CreatedAt.Format("2006-01") != month {
			continue
		}
		result = append(result, f)
	}
	return result, nil
}

func (r *fakeFichaRepository) FindByACSID(acsID uuid.UUID) ([]*domain.Ficha, error) {
	var result []*domain.Ficha
	for _, f := range r.saved {
		if f.ACSID == acsID {
			result = append(result, f)
		}
	}
	return result, nil
}

func (r *fakeFichaRepository) CountByACSID(acsID uuid.UUID) (int, error) {
	fichas, err := r.FindByACSID(acsID)
	return len(fichas), err
}

func (r *fakeFichaRepository) CountByRequestType(start, end string) ([]domain.RequestTypeStat, error) {
	counts := map[string]int{}
	for _, f := range r.saved {
		counts[f.RequestType]++
	}
	var stats []domain.RequestTypeStat
	for requestType, count := range counts {
		stats = append(stats, domain.RequestTypeStat{RequestType: requestType, Count: count})
	}
	return stats, nil
}

func (r *fakeFichaRepository) CountByACS(start, end string) ([]domain.ACSStat, error) {
	counts := map[uuid.UUID]int{}
	for _, f := range r.saved {
		counts[f.ACSID]++
	}
	var stats []domain.ACSStat
	for acsID, count := range counts {
		stats = append(stats, domain.ACSStat{ACSID: acsID, Count: count})
	}
	return stats, nil
}

func (r *fakeFichaRepository) FindMonths() ([]string, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	seen := map[string]bool{}
	var months []string
	for _, f := range r.saved {
		month := f.CreatedAt.Format("2006-01")
		if !seen[month] {
			seen[month] = true
			months = append(months, month)
		}
	}
	return months, nil
}

type fakeACSRepositoryForFicha struct {
	byID map[uuid.UUID]*domain.ACS
}

func newFakeACSRepositoryForFicha(acsID uuid.UUID) *fakeACSRepositoryForFicha {
	return &fakeACSRepositoryForFicha{
		byID: map[uuid.UUID]*domain.ACS{
			acsID: {ID: acsID, Name: "Maria ACS", Phone: "11999999999", CreatedAt: time.Now()},
		},
	}
}

func (r *fakeACSRepositoryForFicha) Save(acs *domain.ACS) error   { return nil }
func (r *fakeACSRepositoryForFicha) Update(acs *domain.ACS) error { return nil }
func (r *fakeACSRepositoryForFicha) Delete(id uuid.UUID) error    { return nil }
func (r *fakeACSRepositoryForFicha) FindAll() ([]*domain.ACS, error) {
	return nil, nil
}
func (r *fakeACSRepositoryForFicha) FindByID(id uuid.UUID) (*domain.ACS, error) {
	return r.byID[id], nil
}
func (r *fakeACSRepositoryForFicha) FindByNameExact(name string) (*domain.ACS, error) {
	return nil, nil
}

type fakeDeleteLogRepository struct {
	records int
	err     error
}

func (r *fakeDeleteLogRepository) Record() error {
	if r.err != nil {
		return r.err
	}
	r.records++
	return nil
}

func newTestFichaService(t *testing.T) (*FichaService, uuid.UUID) {
	t.Helper()
	acsID := uuid.New()
	repo := &fakeFichaRepository{}
	acsRepo := newFakeACSRepositoryForFicha(acsID)
	return NewFichaService(repo, acsRepo, &fakeDeleteLogRepository{}), acsID
}

func TestFichaService_RegisterFicha(t *testing.T) {
	t.Run("registers a valid ficha", func(t *testing.T) {
		service, acsID := newTestFichaService(t)

		ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", acsID, "11988887777", true)

		assert.NoError(t, err)
		assert.NotNil(t, ficha)
		assert.Equal(t, "John Doe", ficha.FullName)
	})

	t.Run("rejects an invalid ficha without persisting it", func(t *testing.T) {
		service, acsID := newTestFichaService(t)

		ficha, err := service.RegisterFicha("", "Exame de sangue", acsID, "", true)

		assert.ErrorIs(t, err, domain.ErrInvalidFullName)
		assert.Nil(t, ficha)
	})

	t.Run("rejects a ficha for an ACS that does not exist", func(t *testing.T) {
		service, _ := newTestFichaService(t)

		ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", uuid.New(), "", true)

		assert.ErrorIs(t, err, domain.ErrACSNotFound)
		assert.Nil(t, ficha)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		acsID := uuid.New()
		repo := &fakeFichaRepository{saveErr: errors.New("disk full")}
		acsRepo := newFakeACSRepositoryForFicha(acsID)
		service := NewFichaService(repo, acsRepo, &fakeDeleteLogRepository{})

		ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)

		assert.Error(t, err)
		assert.Nil(t, ficha)
	})
}

func TestFichaService_UpdateAndDeleteFicha(t *testing.T) {
	service, acsID := newTestFichaService(t)

	ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", acsID, "11988887777", false)
	require.NoError(t, err)

	updated, err := service.UpdateFicha(ficha.ID, "John Updated", "RX", acsID, "11988887777", true)
	require.NoError(t, err)
	assert.Equal(t, "John Updated", updated.FullName)
	assert.True(t, updated.Notified)

	require.NoError(t, service.DeleteFicha(ficha.ID))

	fichas, err := service.ListFichas("", "")
	require.NoError(t, err)
	assert.Empty(t, fichas)
}

func TestFichaService_DeleteFicha_LogsDeletion(t *testing.T) {
	acsID := uuid.New()
	repo := &fakeFichaRepository{}
	acsRepo := newFakeACSRepositoryForFicha(acsID)
	deleteLog := &fakeDeleteLogRepository{}
	service := NewFichaService(repo, acsRepo, deleteLog)

	ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)
	require.NoError(t, err)

	require.NoError(t, service.DeleteFicha(ficha.ID))
	assert.Equal(t, 1, deleteLog.records)
}

func TestFichaService_DeleteFicha_PropagatesDeleteLogError(t *testing.T) {
	acsID := uuid.New()
	repo := &fakeFichaRepository{}
	acsRepo := newFakeACSRepositoryForFicha(acsID)
	deleteLog := &fakeDeleteLogRepository{err: errors.New("disk full")}
	service := NewFichaService(repo, acsRepo, deleteLog)

	ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)
	require.NoError(t, err)

	assert.Error(t, service.DeleteFicha(ficha.ID))
}

func TestFichaService_ListFichas(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)
	_, _ = service.RegisterFicha("Jane Doe", "Exame de sangue", acsID, "", true)

	fichas, err := service.ListFichas("", "")

	assert.NoError(t, err)
	assert.Len(t, fichas, 2)
}

func TestFichaService_ListFichas_FilterByName(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)
	_, _ = service.RegisterFicha("Jane Doe", "Exame de sangue", acsID, "", true)

	fichas, err := service.ListFichas("John Doe", "")

	assert.NoError(t, err)
	assert.Len(t, fichas, 1)
	assert.Equal(t, "John Doe", fichas[0].FullName)
}

func TestFichaService_ListFichas_FilterByMonth(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)

	currentMonth := time.Now().Format("2006-01")

	fichas, err := service.ListFichas("", currentMonth)
	assert.NoError(t, err)
	assert.Len(t, fichas, 1)

	fichas, err = service.ListFichas("", "1999-01")
	assert.NoError(t, err)
	assert.Empty(t, fichas)
}

func TestFichaService_ListFichasByACS(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)

	fichas, err := service.ListFichasByACS(acsID)
	assert.NoError(t, err)
	assert.Len(t, fichas, 1)

	fichas, err = service.ListFichasByACS(uuid.New())
	assert.NoError(t, err)
	assert.Empty(t, fichas)
}

func TestFichaService_ListAvailableMonths(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)

	months, err := service.ListAvailableMonths()

	assert.NoError(t, err)
	assert.Contains(t, months, time.Now().Format("2006-01"))
}

func TestFichaService_StatsByRequestType(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)
	_, _ = service.RegisterFicha("Jane Doe", "Exame de sangue", acsID, "", true)
	_, _ = service.RegisterFicha("Mary Doe", "RX", acsID, "", true)

	stats, err := service.StatsByRequestType("", "")
	require.NoError(t, err)
	require.Len(t, stats, 2)

	byType := map[string]int{}
	for _, s := range stats {
		byType[s.RequestType] = s.Count
	}
	assert.Equal(t, 2, byType["Exame de sangue"])
	assert.Equal(t, 1, byType["RX"])
}

func TestFichaService_StatsByACS(t *testing.T) {
	service, acsID := newTestFichaService(t)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", acsID, "", true)
	_, _ = service.RegisterFicha("Jane Doe", "RX", acsID, "", true)

	stats, err := service.StatsByACS("", "")
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, acsID, stats[0].ACSID)
	assert.Equal(t, 2, stats[0].Count)
}
