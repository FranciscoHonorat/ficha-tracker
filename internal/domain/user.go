package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// User represents a local account used to gate access to the app. There is
// no per-user data isolation: any account can see all ACS and fichas.
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

// NewUser validates the input and creates a new User, timestamped at creation time.
// passwordHash must already be hashed by the caller.
func NewUser(id uuid.UUID, username, passwordHash string) (*User, error) {
	user := &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

// Validate returns an error describing the first invalid field found, if any.
func (u *User) Validate() error {
	if u.ID == uuid.Nil {
		return ErrInvalidID
	}
	if u.Username == "" {
		return ErrInvalidUsername
	}
	if u.PasswordHash == "" {
		return ErrInvalidPassword
	}
	return nil
}
