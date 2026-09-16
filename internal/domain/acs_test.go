package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewACS(t *testing.T) {
	t.Run("creates a valid ACS", func(t *testing.T) {
		id := uuid.New()
		acs, err := NewACS(id, "Maria", "11999999999")

		assert.NoError(t, err)
		assert.NotNil(t, acs)
		assert.Equal(t, id, acs.ID)
		assert.Equal(t, "Maria", acs.Name)
		assert.Equal(t, "11999999999", acs.Phone)
		assert.False(t, acs.CreatedAt.IsZero())
	})

	t.Run("rejects a nil ID", func(t *testing.T) {
		acs, err := NewACS(uuid.Nil, "Maria", "11999999999")
		assert.ErrorIs(t, err, ErrInvalidID)
		assert.Nil(t, acs)
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		acs, err := NewACS(uuid.New(), "", "11999999999")
		assert.ErrorIs(t, err, ErrInvalidACSName)
		assert.Nil(t, acs)
	})

	t.Run("rejects an empty phone", func(t *testing.T) {
		acs, err := NewACS(uuid.New(), "Maria", "")
		assert.ErrorIs(t, err, ErrInvalidACSPhone)
		assert.Nil(t, acs)
	})
}
