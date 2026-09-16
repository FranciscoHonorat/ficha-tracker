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
	ErrInvalidACSID       = errors.New("invalid ACS")
)

// Ficha represents a single print record of a "ficha de saúde":
// the moment a form was printed for a patient by a health agent (ACS).
type Ficha struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"fullName"`
	RequestType string    `json:"requestType"`
	ACSID       uuid.UUID `json:"acsId"`
	// ACSName is a read-only, denormalized display field populated by
	// repository queries that join against the acs table. It is never
	// persisted directly and is ignored by Validate/Equals.
	ACSName   string    `json:"acsName,omitempty"`
	Phone     string    `json:"phone"`
	Notified  bool      `json:"notified"`
	CreatedAt time.Time `json:"createdAt"`
}

// NewFicha validates the input and creates a new Ficha, timestamped at creation time.
func NewFicha(id uuid.UUID, fullName, requestType string, acsID uuid.UUID, phone string, notified bool) (*Ficha, error) {
	ficha := &Ficha{
		ID:          id,
		FullName:    fullName,
		RequestType: requestType,
		ACSID:       acsID,
		Phone:       phone,
		Notified:    notified,
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
	if f.ACSID == uuid.Nil {
		return ErrInvalidACSID
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
		f.ACSID == other.ACSID &&
		f.Phone == other.Phone &&
		f.Notified == other.Notified
}

func (f *Ficha) Marshal() map[string]any {
	return map[string]any{
		"id":           f.ID.String(),
		"full_name":    f.FullName,
		"request_type": f.RequestType,
		"acs_id":       f.ACSID.String(),
		"phone":        f.Phone,
		"notified":     f.Notified,
		"created_at":   f.CreatedAt.Format(time.RFC3339),
	}
}

func (f *Ficha) Unmarshal(data map[string]any) error {
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

	acsIDStr, ok := data["acs_id"].(string)
	if !ok {
		return ErrInvalidACSID
	}
	acsID, err := uuid.Parse(acsIDStr)
	if err != nil {
		return ErrInvalidACSID
	}

	phone, _ := data["phone"].(string)
	notified, _ := data["notified"].(bool)

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
	f.ACSID = acsID
	f.Phone = phone
	f.Notified = notified
	f.CreatedAt = createdAt
	return nil
}
