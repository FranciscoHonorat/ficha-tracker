package service

import (
	"errors"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

// FichaRepository is the persistence port required by FichaService.
// It is implemented by the SQLite repository in internal/repository.
type FichaRepository interface {
	Save(ficha *domain.Ficha) error
	FindAll() ([]*domain.Ficha, error)
	FindByName(name string) ([]*domain.Ficha, error)
}

type FichaService struct {
	repo FichaRepository
}

func NewFichaService(repo FichaRepository) *FichaService {
	return &FichaService{repo: repo}
}

// RegisterFicha validates and persists a new ficha print record.
func (s *FichaService) RegisterFicha(fullName, requestType, acs string) (*domain.Ficha, error) {
	ficha, err := domain.NewFicha(uuid.New(), fullName, requestType, acs)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ficha); err != nil {
		return nil, err
	}

	return ficha, nil
}

// ListFichas returns every ficha ever registered, most recent first.
func (s *FichaService) ListFichas() ([]*domain.Ficha, error) {
	return s.repo.FindAll()
}

// SearchFichasByName returns every ficha whose name matches the given query.
func (s *FichaService) SearchFichasByName(name string) ([]*domain.Ficha, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	return s.repo.FindByName(name)
}
