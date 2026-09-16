package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidACSName   = errors.New("invalid ACS name")
	ErrInvalidACSPhone  = errors.New("invalid ACS phone")
	ErrDuplicateACSName = errors.New("ACS with this name already exists")
	ErrACSHasFichas     = errors.New("ACS still has fichas registered")
	ErrACSNotFound      = errors.New("ACS not found")
)

// ACS represents a community health agent ("Agente Comunitário de Saúde") who
// requests fichas on behalf of patients.
type ACS struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"createdAt"`
}

// NewACS validates the input and creates a new ACS, timestamped at creation time.
func NewACS(id uuid.UUID, name, phone string) (*ACS, error) {
	acs := &ACS{
		ID:        id,
		Name:      name,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
	if err := acs.Validate(); err != nil {
		return nil, err
	}
	return acs, nil
}

// Validate returns an error describing the first invalid field found, if any.
func (a *ACS) Validate() error {
	if a.ID == uuid.Nil {
		return ErrInvalidID
	}
	if a.Name == "" {
		return ErrInvalidACSName
	}
	if a.Phone == "" {
		return ErrInvalidACSPhone
	}
	return nil
}
