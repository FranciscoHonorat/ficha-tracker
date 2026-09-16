package service

import (
	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

// FichaRepository is the persistence port required by FichaService.
// It is implemented by the SQLite repository in internal/repository.
type FichaRepository interface {
	Save(ficha *domain.Ficha) error
	Update(ficha *domain.Ficha) error
	Delete(id uuid.UUID) error
	// Find returns fichas matching an optional name filter (substring) and
	// an optional month filter ("YYYY-MM"), most recent first. An empty
	// string skips that filter.
	Find(name, month string) ([]*domain.Ficha, error)
	// FindByACSID returns every ficha registered for the given ACS, most
	// recent first.
	FindByACSID(acsID uuid.UUID) ([]*domain.Ficha, error)
	// CountByACSID returns how many fichas reference the given ACS.
	CountByACSID(acsID uuid.UUID) (int, error)
	// FindMonths returns every month ("YYYY-MM") that has at least one
	// ficha registered, most recent first.
	FindMonths() ([]string, error)
	// CountByRequestType returns, for the given analysis window ("YYYY-MM-DD"
	// dates, either bound may be empty to leave it open), how many fichas
	// were registered per request type, most frequent first.
	CountByRequestType(start, end string) ([]domain.RequestTypeStat, error)
	// CountByACS returns, for the given analysis window ("YYYY-MM-DD" dates,
	// either bound may be empty to leave it open), how many fichas were
	// registered per ACS, most frequent first.
	CountByACS(start, end string) ([]domain.ACSStat, error)
}

type FichaService struct {
	repo      FichaRepository
	acsRepo   ACSRepository
	deleteLog DeleteLogRepository
}

func NewFichaService(repo FichaRepository, acsRepo ACSRepository, deleteLog DeleteLogRepository) *FichaService {
	return &FichaService{repo: repo, acsRepo: acsRepo, deleteLog: deleteLog}
}

// RegisterFicha validates and persists a new ficha print record.
func (s *FichaService) RegisterFicha(fullName, requestType string, acsID uuid.UUID, phone string, notified bool) (*domain.Ficha, error) {
	if err := s.ensureACSExists(acsID); err != nil {
		return nil, err
	}

	ficha, err := domain.NewFicha(uuid.New(), fullName, requestType, acsID, phone, notified)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ficha); err != nil {
		return nil, err
	}

	return ficha, nil
}

// UpdateFicha validates and persists changes to an existing ficha.
func (s *FichaService) UpdateFicha(id uuid.UUID, fullName, requestType string, acsID uuid.UUID, phone string, notified bool) (*domain.Ficha, error) {
	if err := s.ensureACSExists(acsID); err != nil {
		return nil, err
	}

	ficha, err := domain.NewFicha(id, fullName, requestType, acsID, phone, notified)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ficha); err != nil {
		return nil, err
	}

	return ficha, nil
}

// DeleteFicha removes a ficha by id and logs the deletion's timestamp.
func (s *FichaService) DeleteFicha(id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return s.deleteLog.Record()
}

// ListFichas returns fichas matching an optional name filter and an
// optional month filter ("YYYY-MM"), most recent first. Pass an empty
// string to skip a filter.
func (s *FichaService) ListFichas(name, month string) ([]*domain.Ficha, error) {
	return s.repo.Find(name, month)
}

// ListFichasByACS returns every ficha registered for the given ACS, most
// recent first.
func (s *FichaService) ListFichasByACS(acsID uuid.UUID) ([]*domain.Ficha, error) {
	return s.repo.FindByACSID(acsID)
}

// ListAvailableMonths returns every month ("YYYY-MM") that has at least one
// ficha registered, most recent first.
func (s *FichaService) ListAvailableMonths() ([]string, error) {
	return s.repo.FindMonths()
}

// StatsByRequestType returns exam counts per request type for the given
// analysis window ("YYYY-MM-DD" dates, either bound may be empty).
func (s *FichaService) StatsByRequestType(start, end string) ([]domain.RequestTypeStat, error) {
	return s.repo.CountByRequestType(start, end)
}

// StatsByACS returns exam counts per ACS for the given analysis window
// ("YYYY-MM-DD" dates, either bound may be empty).
func (s *FichaService) StatsByACS(start, end string) ([]domain.ACSStat, error) {
	return s.repo.CountByACS(start, end)
}

func (s *FichaService) ensureACSExists(acsID uuid.UUID) error {
	acs, err := s.acsRepo.FindByID(acsID)
	if err != nil {
		return err
	}
	if acs == nil {
		return domain.ErrACSNotFound
	}
	return nil
}
