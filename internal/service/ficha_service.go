package service

import (
	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

// FichaRepository is the persistence port required by FichaService.
// It is implemented by the SQLite repository in internal/repository.
type FichaRepository interface {
	Save(ficha *domain.Ficha) error
	// Find returns fichas matching an optional name filter (substring) and
	// an optional month filter ("YYYY-MM"), most recent first. An empty
	// string skips that filter.
	Find(name, month string) ([]*domain.Ficha, error)
	// FindMonths returns every month ("YYYY-MM") that has at least one
	// ficha registered, most recent first.
	FindMonths() ([]string, error)
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

// ListFichas returns fichas matching an optional name filter and an
// optional month filter ("YYYY-MM"), most recent first. Pass an empty
// string to skip a filter.
func (s *FichaService) ListFichas(name, month string) ([]*domain.Ficha, error) {
	return s.repo.Find(name, month)
}

// ListAvailableMonths returns every month ("YYYY-MM") that has at least one
// ficha registered, most recent first.
func (s *FichaService) ListAvailableMonths() ([]string, error) {
	return s.repo.FindMonths()
}
