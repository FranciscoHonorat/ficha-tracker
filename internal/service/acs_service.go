package service

import (
	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

// ACSRepository is the persistence port required by ACSService and
// FichaService. It is implemented by the SQLite repository in
// internal/repository.
type ACSRepository interface {
	Save(acs *domain.ACS) error
	Update(acs *domain.ACS) error
	Delete(id uuid.UUID) error
	FindAll() ([]*domain.ACS, error)
	FindByID(id uuid.UUID) (*domain.ACS, error)
	// FindByNameExact returns the ACS with the given name, matched
	// case-insensitively, or nil if none exists.
	FindByNameExact(name string) (*domain.ACS, error)
}

// FichaCounter is the narrow slice of FichaRepository that ACSService needs
// to refuse deleting an ACS that still has fichas registered against it.
type FichaCounter interface {
	CountByACSID(acsID uuid.UUID) (int, error)
}

type ACSService struct {
	repo         ACSRepository
	fichaCounter FichaCounter
}

func NewACSService(repo ACSRepository, fichaCounter FichaCounter) *ACSService {
	return &ACSService{repo: repo, fichaCounter: fichaCounter}
}

// RegisterACS validates and persists a new ACS, rejecting duplicate names
// (case-insensitive).
func (s *ACSService) RegisterACS(name, phone string) (*domain.ACS, error) {
	if err := s.ensureNameAvailable(name, uuid.Nil); err != nil {
		return nil, err
	}

	acs, err := domain.NewACS(uuid.New(), name, phone)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(acs); err != nil {
		return nil, err
	}

	return acs, nil
}

// UpdateACS validates and persists changes to an existing ACS, rejecting
// duplicate names (case-insensitive) held by a different ACS.
func (s *ACSService) UpdateACS(id uuid.UUID, name, phone string) (*domain.ACS, error) {
	if err := s.ensureNameAvailable(name, id); err != nil {
		return nil, err
	}

	acs, err := domain.NewACS(id, name, phone)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(acs); err != nil {
		return nil, err
	}

	return acs, nil
}

// DeleteACS removes an ACS, refusing to do so while fichas still reference it.
func (s *ACSService) DeleteACS(id uuid.UUID) error {
	count, err := s.fichaCounter.CountByACSID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrACSHasFichas
	}
	return s.repo.Delete(id)
}

// ListACS returns every registered ACS.
func (s *ACSService) ListACS() ([]*domain.ACS, error) {
	return s.repo.FindAll()
}

func (s *ACSService) ensureNameAvailable(name string, excludeID uuid.UUID) error {
	existing, err := s.repo.FindByNameExact(name)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return domain.ErrDuplicateACSName
	}
	return nil
}
