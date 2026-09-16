package service

import (
	"testing"

	"ficha-tracker/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUserRepository struct {
	saved []*domain.User
}

func (r *fakeUserRepository) Save(user *domain.User) error {
	r.saved = append(r.saved, user)
	return nil
}

func (r *fakeUserRepository) FindByUsername(username string) (*domain.User, error) {
	for _, u := range r.saved {
		if equalFold(u.Username, username) {
			return u, nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepository) CountAll() (int, error) {
	return len(r.saved), nil
}

func TestAuthService_HasAccount(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewAuthService(repo)

	has, err := service.HasAccount()
	require.NoError(t, err)
	assert.False(t, has)

	_, err = service.CreateAccount("admin", "s3cret")
	require.NoError(t, err)

	has, err = service.HasAccount()
	require.NoError(t, err)
	assert.True(t, has)
}

func TestAuthService_CreateAccount(t *testing.T) {
	t.Run("creates a valid account", func(t *testing.T) {
		repo := &fakeUserRepository{}
		service := NewAuthService(repo)

		user, err := service.CreateAccount("admin", "s3cret")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEqual(t, "s3cret", user.PasswordHash)
	})

	t.Run("rejects a duplicate username case-insensitively", func(t *testing.T) {
		repo := &fakeUserRepository{}
		service := NewAuthService(repo)

		_, err := service.CreateAccount("admin", "s3cret")
		require.NoError(t, err)

		_, err = service.CreateAccount("Admin", "other")
		assert.ErrorIs(t, err, domain.ErrUsernameTaken)
	})

	t.Run("rejects an empty password", func(t *testing.T) {
		repo := &fakeUserRepository{}
		service := NewAuthService(repo)

		_, err := service.CreateAccount("admin", "")
		assert.ErrorIs(t, err, domain.ErrInvalidPassword)
	})
}

func TestAuthService_Login(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewAuthService(repo)
	_, err := service.CreateAccount("admin", "s3cret")
	require.NoError(t, err)

	t.Run("accepts correct credentials", func(t *testing.T) {
		assert.NoError(t, service.Login("admin", "s3cret"))
	})

	t.Run("is case-insensitive on username", func(t *testing.T) {
		assert.NoError(t, service.Login("Admin", "s3cret"))
	})

	t.Run("rejects a wrong password", func(t *testing.T) {
		assert.ErrorIs(t, service.Login("admin", "wrong"), domain.ErrInvalidCredentials)
	})

	t.Run("rejects an unknown username", func(t *testing.T) {
		assert.ErrorIs(t, service.Login("nobody", "s3cret"), domain.ErrInvalidCredentials)
	})
}
