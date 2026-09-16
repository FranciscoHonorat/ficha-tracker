package service

import (
	"testing"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
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

func (r *fakeUserRepository) UpdatePasswordHash(id uuid.UUID, passwordHash string) error {
	for _, u := range r.saved {
		if u.ID == id {
			u.PasswordHash = passwordHash
			return nil
		}
	}
	return nil
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

func TestAuthService_ChangePassword(t *testing.T) {
	newService := func(t *testing.T) *AuthService {
		t.Helper()
		repo := &fakeUserRepository{}
		service := NewAuthService(repo)
		_, err := service.CreateAccount("admin", "s3cret")
		require.NoError(t, err)
		return service
	}

	t.Run("changes the password with correct credentials", func(t *testing.T) {
		service := newService(t)

		require.NoError(t, service.ChangePassword("admin", "s3cret", "newpass"))
		assert.ErrorIs(t, service.Login("admin", "s3cret"), domain.ErrInvalidCredentials)
		assert.NoError(t, service.Login("admin", "newpass"))
	})

	t.Run("rejects the wrong current password", func(t *testing.T) {
		service := newService(t)

		err := service.ChangePassword("admin", "wrong", "newpass")
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		assert.NoError(t, service.Login("admin", "s3cret"))
	})

	t.Run("rejects an empty new password", func(t *testing.T) {
		service := newService(t)

		err := service.ChangePassword("admin", "s3cret", "")
		assert.ErrorIs(t, err, domain.ErrInvalidPassword)
	})

	t.Run("rejects an unknown username", func(t *testing.T) {
		service := newService(t)

		err := service.ChangePassword("nobody", "s3cret", "newpass")
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}
