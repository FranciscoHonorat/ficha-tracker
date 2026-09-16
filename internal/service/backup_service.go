package service

import "errors"

var ErrBackupPathRequired = errors.New("backup destination path is required")

// BackupRepository writes a consistent snapshot of the database to a path.
// It is implemented by the SQLite repository in internal/repository.
type BackupRepository interface {
	Backup(destPath string) error
}

type BackupService struct {
	repo BackupRepository
}

func NewBackupService(repo BackupRepository) *BackupService {
	return &BackupService{repo: repo}
}

// Backup writes a consistent snapshot of the database to destPath.
func (s *BackupService) Backup(destPath string) error {
	if destPath == "" {
		return ErrBackupPathRequired
	}
	return s.repo.Backup(destPath)
}
