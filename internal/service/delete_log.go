package service

// DeleteLogRepository records that a ficha or ACS deletion happened, storing
// nothing about what was deleted, for a 90-day retention audit trail. It is
// implemented by the SQLite repository in internal/repository.
type DeleteLogRepository interface {
	Record() error
}
