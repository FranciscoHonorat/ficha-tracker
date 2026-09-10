package service

import (
	"errors"
	"testing"

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

func (r *fakeFichaRepository) FindAll() ([]*domain.Ficha, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.saved, nil
}

func (r *fakeFichaRepository) FindByName(name string) ([]*domain.Ficha, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	var result []*domain.Ficha
	for _, f := range r.saved {
		if f.FullName == name {
			result = append(result, f)
		}
	}
	return result, nil
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

	fichas, err := service.ListFichas()

	assert.NoError(t, err)
	assert.Len(t, fichas, 2)
}

func TestFichaService_SearchFichasByName(t *testing.T) {
	repo := &fakeFichaRepository{}
	service := NewFichaService(repo)

	_, _ = service.RegisterFicha("John Doe", "Exame de sangue", "Maria ACS")
	_, _ = service.RegisterFicha("Jane Doe", "Exame de sangue", "Maria ACS")

	t.Run("finds fichas matching the name", func(t *testing.T) {
		fichas, err := service.SearchFichasByName("John Doe")

		assert.NoError(t, err)
		assert.Len(t, fichas, 1)
		assert.Equal(t, "John Doe", fichas[0].FullName)
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		fichas, err := service.SearchFichasByName("")

		assert.Error(t, err)
		assert.Nil(t, fichas)
	})
}
