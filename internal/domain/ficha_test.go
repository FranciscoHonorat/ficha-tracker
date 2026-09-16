package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewFicha(t *testing.T) {
	t.Run("creates a valid ficha", func(t *testing.T) {
		id := uuid.New()
		acsID := uuid.New()
		ficha, err := NewFicha(id, "John Doe", "Exame de sangue", acsID, "11999999999", true)

		assert.NoError(t, err)
		assert.NotNil(t, ficha)
		assert.Equal(t, id, ficha.ID)
		assert.Equal(t, "John Doe", ficha.FullName)
		assert.Equal(t, "Exame de sangue", ficha.RequestType)
		assert.Equal(t, acsID, ficha.ACSID)
		assert.Equal(t, "11999999999", ficha.Phone)
		assert.True(t, ficha.Notified)
		assert.False(t, ficha.CreatedAt.IsZero())
	})

	t.Run("rejects a nil ID", func(t *testing.T) {
		ficha, err := NewFicha(uuid.Nil, "John Doe", "Exame de sangue", uuid.New(), "", true)
		assert.ErrorIs(t, err, ErrInvalidID)
		assert.Nil(t, ficha)
	})

	t.Run("rejects an empty full name", func(t *testing.T) {
		ficha, err := NewFicha(uuid.New(), "", "Exame de sangue", uuid.New(), "", true)
		assert.ErrorIs(t, err, ErrInvalidFullName)
		assert.Nil(t, ficha)
	})

	t.Run("rejects an empty request type", func(t *testing.T) {
		ficha, err := NewFicha(uuid.New(), "John Doe", "", uuid.New(), "", true)
		assert.ErrorIs(t, err, ErrInvalidRequestType)
		assert.Nil(t, ficha)
	})

	t.Run("rejects a nil ACS id", func(t *testing.T) {
		ficha, err := NewFicha(uuid.New(), "John Doe", "Exame de sangue", uuid.Nil, "", true)
		assert.ErrorIs(t, err, ErrInvalidACSID)
		assert.Nil(t, ficha)
	})
}

func TestFicha_IsValid(t *testing.T) {
	ficha, err := NewFicha(uuid.New(), "John Doe", "Exame de sangue", uuid.New(), "", true)
	assert.NoError(t, err)
	assert.True(t, ficha.IsValid())

	ficha.RequestType = ""
	assert.False(t, ficha.IsValid())
}

func TestFicha_Equals(t *testing.T) {
	id := uuid.New()
	acsID := uuid.New()
	a, _ := NewFicha(id, "John Doe", "Exame de sangue", acsID, "", true)
	b, _ := NewFicha(id, "John Doe", "Exame de sangue", acsID, "", true)
	c, _ := NewFicha(uuid.New(), "Jane Doe", "Exame de sangue", acsID, "", true)

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
	assert.False(t, a.Equals(nil))
}

func TestFicha_MarshalUnmarshal(t *testing.T) {
	original, err := NewFicha(uuid.New(), "John Doe", "Exame de sangue", uuid.New(), "11999999999", false)
	assert.NoError(t, err)

	data := original.Marshal()

	roundTripped := &Ficha{}
	err = roundTripped.Unmarshal(data)

	assert.NoError(t, err)
	assert.True(t, original.Equals(roundTripped))
	assert.WithinDuration(t, original.CreatedAt, roundTripped.CreatedAt, time.Second)
}
