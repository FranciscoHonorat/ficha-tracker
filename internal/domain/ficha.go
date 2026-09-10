package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidID          = errors.New("invalid ID")
	ErrInvalidFullName    = errors.New("invalid full name")
	ErrInvalidRequestType = errors.New("invalid request type")
	ErrInvalidACS         = errors.New("invalid ACS")
)

// Ficha represents a single print record of a "ficha de saúde":
// the moment a form was printed for a patient by a health agent (ACS).
type Ficha struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"fullName"`
	RequestType string    `json:"requestType"`
	ACS         string    `json:"acs"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewFicha validates the input and creates a new Ficha, timestamped at creation time.
func NewFicha(id uuid.UUID, fullName, requestType, acs string) (*Ficha, error) {
	ficha := &Ficha{
		ID:          id,
		FullName:    fullName,
		RequestType: requestType,
		ACS:         acs,
		CreatedAt:   time.Now(),
	}
	if err := ficha.Validate(); err != nil {
		return nil, err
	}
	return ficha, nil
}

// Validate returns an error describing the first invalid field found, if any.
func (f *Ficha) Validate() error {
	if f.ID == uuid.Nil {
		return ErrInvalidID
	}
	if f.FullName == "" {
		return ErrInvalidFullName
	}
	if f.RequestType == "" {
		return ErrInvalidRequestType
	}
	if f.ACS == "" {
		return ErrInvalidACS
	}
	return nil
}

func (f *Ficha) IsValid() bool {
	return f.Validate() == nil
}

func (f *Ficha) Equals(other *Ficha) bool {
	if f == nil || other == nil {
		return false
	}
	return f.ID == other.ID &&
		f.FullName == other.FullName &&
		f.RequestType == other.RequestType &&
		f.ACS == other.ACS
}

func (f *Ficha) Marshal() map[string]interface{} {
	return map[string]interface{}{
		"id":           f.ID.String(),
		"full_name":    f.FullName,
		"request_type": f.RequestType,
		"acs":          f.ACS,
		"created_at":   f.CreatedAt.Format(time.RFC3339),
	}
}

func (f *Ficha) Unmarshal(data map[string]interface{}) error {
	idStr, ok := data["id"].(string)
	if !ok {
		return ErrInvalidID
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ErrInvalidID
	}

	fullName, ok := data["full_name"].(string)
	if !ok {
		return ErrInvalidFullName
	}

	requestType, ok := data["request_type"].(string)
	if !ok {
		return ErrInvalidRequestType
	}

	acs, ok := data["acs"].(string)
	if !ok {
		return ErrInvalidACS
	}

	createdAt := time.Now()
	if createdAtStr, ok := data["created_at"].(string); ok {
		parsed, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return errors.New("invalid created_at format")
		}
		createdAt = parsed
	}

	f.ID = id
	f.FullName = fullName
	f.RequestType = requestType
	f.ACS = acs
	f.CreatedAt = createdAt
	return nil
}
