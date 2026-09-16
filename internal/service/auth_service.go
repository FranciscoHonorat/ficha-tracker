package service

import (
	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository is the persistence port required by AuthService. It is
// implemented by the SQLite repository in internal/repository.
type UserRepository interface {
	Save(user *domain.User) error
	// FindByUsername returns the user with the given username, matched
	// case-insensitively, or nil if none exists.
	FindByUsername(username string) (*domain.User, error)
	CountAll() (int, error)
	// UpdatePasswordHash replaces the stored password hash for the given user.
	UpdatePasswordHash(id uuid.UUID, passwordHash string) error
}

// AuthService gates access to the app with a single, simple local account
// system: no per-user data isolation, just a login screen.
type AuthService struct {
	repo UserRepository
}

func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// HasAccount reports whether any account has been created yet.
func (s *AuthService) HasAccount() (bool, error) {
	count, err := s.repo.CountAll()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateAccount validates and persists a new account, rejecting a username
// that is already taken (case-insensitive).
func (s *AuthService) CreateAccount(username, password string) (*domain.User, error) {
	if password == "" {
		return nil, domain.ErrInvalidPassword
	}

	existing, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(uuid.New(), username, string(hash))
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login verifies the given credentials against the stored account.
func (s *AuthService) Login(username, password string) error {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.ErrInvalidCredentials
	}

	return nil
}

// ChangePassword verifies the current credentials and replaces the account's
// password.
func (s *AuthService) ChangePassword(username, oldPassword, newPassword string) error {
	if newPassword == "" {
		return domain.ErrInvalidPassword
	}

	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePasswordHash(user.ID, string(hash))
}
