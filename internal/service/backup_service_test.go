package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeBackupRepository struct {
	backedUpTo string
	err        error
}

func (r *fakeBackupRepository) Backup(destPath string) error {
	if r.err != nil {
		return r.err
	}
	r.backedUpTo = destPath
	return nil
}

func TestBackupService_Backup(t *testing.T) {
	t.Run("delegates to the repository", func(t *testing.T) {
		repo := &fakeBackupRepository{}
		service := NewBackupService(repo)

		assert.NoError(t, service.Backup("/tmp/backup.db"))
		assert.Equal(t, "/tmp/backup.db", repo.backedUpTo)
	})

	t.Run("rejects an empty path", func(t *testing.T) {
		repo := &fakeBackupRepository{}
		service := NewBackupService(repo)

		assert.ErrorIs(t, service.Backup(""), ErrBackupPathRequired)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := &fakeBackupRepository{err: errors.New("disk full")}
		service := NewBackupService(repo)

		assert.Error(t, service.Backup("/tmp/backup.db"))
	})
}
