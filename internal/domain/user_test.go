package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("creates a valid user", func(t *testing.T) {
		id := uuid.New()
		user, err := NewUser(id, "admin", "hashed-password")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, "hashed-password", user.PasswordHash)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("rejects a nil ID", func(t *testing.T) {
		user, err := NewUser(uuid.Nil, "admin", "hashed-password")
		assert.ErrorIs(t, err, ErrInvalidID)
		assert.Nil(t, user)
	})

	t.Run("rejects an empty username", func(t *testing.T) {
		user, err := NewUser(uuid.New(), "", "hashed-password")
		assert.ErrorIs(t, err, ErrInvalidUsername)
		assert.Nil(t, user)
	})

	t.Run("rejects an empty password hash", func(t *testing.T) {
		user, err := NewUser(uuid.New(), "admin", "")
		assert.ErrorIs(t, err, ErrInvalidPassword)
		assert.Nil(t, user)
	})
}
