package service

import (
	"errors"
	"testing"
	"time"

	"ficha-tracker/internal/domain"

	"github.com/stretchr/testify/assert"
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

func TestFichaService_RegisterFicha(t *testing.T) {
	t.Run("registers a valid ficha", func(t *testing.T) {
		repo := &fakeFichaRepository{}
		service := NewFichaService(repo)

		ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")

		assert.NoError(t, err)
		assert.NotNil(t, ficha)
		assert.Equal(t, "John Doe", ficha.FullName)
		assert.Len(t, repo.saved, 1)
	})

	t.Run("rejects an invalid ficha without persisting it", func(t *testing.T) {
		repo := &fakeFichaRepository{}
		service := NewFichaService(repo)

		ficha, err := service.RegisterFicha("", "Exame de sangue", "Maria ACS")

		assert.ErrorIs(t, err, domain.ErrInvalidFullName)
		assert.Nil(t, ficha)
		assert.Empty(t, repo.saved)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := &fakeFichaRepository{saveErr: errors.New("disk full")}
		service := NewFichaService(repo)

		ficha, err := service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")

		assert.Error(t, err)
		assert.Nil(t, ficha)
	})
}

func TestFichaService_ListFichas(t *testing.T) {
	repo := &fakeFichaRepository{}
	service := NewFichaService(repo)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")
	_, _ = service.RegisterFicha("Jane Doe", "Exame de sangue", "Maria ACS")

	fichas, err := service.ListFichas("", "")

	assert.NoError(t, err)
	assert.Len(t, fichas, 2)
}

func TestFichaService_ListFichas_FilterByName(t *testing.T) {
	repo := &fakeFichaRepository{}
	service := NewFichaService(repo)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")
	_, _ = service.RegisterFicha("Jane Doe", "Exame de sangue", "Maria ACS")

	fichas, err := service.ListFichas("John Doe", "")

	assert.NoError(t, err)
	assert.Len(t, fichas, 1)
	assert.Equal(t, "John Doe", fichas[0].FullName)
}

func TestFichaService_ListFichas_FilterByMonth(t *testing.T) {
	repo := &fakeFichaRepository{}
	service := NewFichaService(repo)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")

	currentMonth := time.Now().Format("2006-01")

	fichas, err := service.ListFichas("", currentMonth)
	assert.NoError(t, err)
	assert.Len(t, fichas, 1)

	fichas, err = service.ListFichas("", "1999-01")
	assert.NoError(t, err)
	assert.Empty(t, fichas)
}

func TestFichaService_ListAvailableMonths(t *testing.T) {
	repo := &fakeFichaRepository{}
	service := NewFichaService(repo)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")

	months, err := service.ListAvailableMonths()

	assert.NoError(t, err)
	assert.Contains(t, months, time.Now().Format("2006-01"))
}
