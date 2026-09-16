package repository

import (
	"database/sql"
	"fmt"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// OpenDB opens (creating if needed) the SQLite database at path and applies schema.
func OpenDB(path string, schema string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	return db, nil
}

// SQLiteFichaRepository persists Fichas in a SQLite database. It implements
// service.FichaRepository.
type SQLiteFichaRepository struct {
	db *sql.DB
}

func NewSQLiteFichaRepository(db *sql.DB) *SQLiteFichaRepository {
	return &SQLiteFichaRepository{db: db}
}

func (r *SQLiteFichaRepository) Save(ficha *domain.Ficha) error {
	// The legacy "acs" text column is unused since v2.0, but on a database
	// migrated from before v2.0 it still carries its original NOT NULL
	// constraint with no default (ALTER TABLE ADD COLUMN cannot retrofit one
	// onto a pre-existing column), so it must always be given a value here.
	_, err := r.db.Exec(
		`INSERT INTO fichas (id, full_name, request_type, acs, acs_id, phone, notified, created_at) VALUES (?, ?, ?, '', ?, ?, ?, ?)`,
		ficha.ID.String(), ficha.FullName, ficha.RequestType, ficha.ACSID.String(), ficha.Phone, ficha.Notified, ficha.CreatedAt.Format(rfc3339),
	)
	if err != nil {
		return fmt.Errorf("saving ficha: %w", err)
	}
	return nil
}

// Update persists changes to an existing ficha.
func (r *SQLiteFichaRepository) Update(ficha *domain.Ficha) error {
	_, err := r.db.Exec(
		`UPDATE fichas SET full_name = ?, request_type = ?, acs_id = ?, phone = ?, notified = ? WHERE id = ?`,
		ficha.FullName, ficha.RequestType, ficha.ACSID.String(), ficha.Phone, ficha.Notified, ficha.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("updating ficha: %w", err)
	}
	return nil
}

// Delete removes a ficha by id.
func (r *SQLiteFichaRepository) Delete(id uuid.UUID) error {
	if _, err := r.db.Exec(`DELETE FROM fichas WHERE id = ?`, id.String()); err != nil {
		return fmt.Errorf("deleting ficha: %w", err)
	}
	return nil
}

// Find returns fichas matching an optional name filter (substring) and an
// optional month filter ("YYYY-MM"), most recent first. An empty string
// skips that filter.
func (r *SQLiteFichaRepository) Find(name, month string) ([]*domain.Ficha, error) {
	query := `SELECT f.id, f.full_name, f.request_type, f.acs_id, f.phone, f.notified, f.created_at, a.name
		FROM fichas f LEFT JOIN acs a ON a.id = f.acs_id WHERE 1=1`
	var args []any

	if name != "" {
		query += ` AND f.full_name LIKE '%' || ? || '%'`
		args = append(args, name)
	}
	if month != "" {
		query += ` AND substr(f.created_at, 1, 7) = ?`
		args = append(args, month)
	}
	query += ` ORDER BY f.created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing fichas: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows, true)
}

// FindByACSID returns every ficha registered for the given ACS, most recent first.
func (r *SQLiteFichaRepository) FindByACSID(acsID uuid.UUID) ([]*domain.Ficha, error) {
	rows, err := r.db.Query(
		`SELECT f.id, f.full_name, f.request_type, f.acs_id, f.phone, f.notified, f.created_at, a.name
		FROM fichas f LEFT JOIN acs a ON a.id = f.acs_id
		WHERE f.acs_id = ? ORDER BY f.created_at DESC`,
		acsID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("listing fichas by acs: %w", err)
	}
	defer rows.Close()

	return scanFichas(rows, true)
}

// CountByACSID returns how many fichas reference the given ACS.
func (r *SQLiteFichaRepository) CountByACSID(acsID uuid.UUID) (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM fichas WHERE acs_id = ?`, acsID.String()).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting fichas by acs: %w", err)
	}
	return count, nil
}

// dateRangeClause builds a "created_at" date-window WHERE fragment (against
// the given column) for an optional "YYYY-MM-DD" start/end, either of which
// may be empty to leave that bound open. It returns the SQL fragment
// (starting with " AND ...", empty if both bounds are empty) and its args.
func dateRangeClause(column, start, end string) (string, []any) {
	var clause string
	var args []any
	if start != "" {
		clause += fmt.Sprintf(" AND substr(%s, 1, 10) >= ?", column)
		args = append(args, start)
	}
	if end != "" {
		clause += fmt.Sprintf(" AND substr(%s, 1, 10) <= ?", column)
		args = append(args, end)
	}
	return clause, args
}

// CountByRequestType returns, for the given analysis window, how many
// fichas were registered per request type, most frequent first.
func (r *SQLiteFichaRepository) CountByRequestType(start, end string) ([]domain.RequestTypeStat, error) {
	clause, args := dateRangeClause("created_at", start, end)
	rows, err := r.db.Query(
		`SELECT request_type, COUNT(*) FROM fichas WHERE 1=1`+clause+` GROUP BY request_type ORDER BY COUNT(*) DESC, request_type ASC`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("counting fichas by request type: %w", err)
	}
	defer rows.Close()

	var stats []domain.RequestTypeStat
	for rows.Next() {
		var stat domain.RequestTypeStat
		if err := rows.Scan(&stat.RequestType, &stat.Count); err != nil {
			return nil, fmt.Errorf("scanning request type stat: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading request type stats: %w", err)
	}
	return stats, nil
}

// CountByACS returns, for the given analysis window, how many fichas were
// registered per ACS, most frequent first.
func (r *SQLiteFichaRepository) CountByACS(start, end string) ([]domain.ACSStat, error) {
	clause, args := dateRangeClause("f.created_at", start, end)
	rows, err := r.db.Query(
		`SELECT f.acs_id, a.name, COUNT(*) FROM fichas f LEFT JOIN acs a ON a.id = f.acs_id
		WHERE 1=1`+clause+` GROUP BY f.acs_id ORDER BY COUNT(*) DESC, a.name ASC`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("counting fichas by acs: %w", err)
	}
	defer rows.Close()

	var stats []domain.ACSStat
	for rows.Next() {
		var (
			acsIDStr string
			acsName  sql.NullString
			count    int
		)
		if err := rows.Scan(&acsIDStr, &acsName, &count); err != nil {
			return nil, fmt.Errorf("scanning acs stat: %w", err)
		}
		acsID, err := uuid.Parse(acsIDStr)
		if err != nil {
			return nil, fmt.Errorf("parsing acs stat id: %w", err)
		}
		stats = append(stats, domain.ACSStat{ACSID: acsID, ACSName: acsName.String, Count: count})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading acs stats: %w", err)
	}
	return stats, nil
}

// FindMonths returns every month ("YYYY-MM") that has at least one ficha
// registered, most recent first.
func (r *SQLiteFichaRepository) FindMonths() ([]string, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT substr(created_at, 1, 7) FROM fichas ORDER BY 1 DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing ficha months: %w", err)
	}
	defer rows.Close()

	var months []string
	for rows.Next() {
		var month string
		if err := rows.Scan(&month); err != nil {
			return nil, fmt.Errorf("scanning ficha month: %w", err)
		}
		months = append(months, month)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading ficha months: %w", err)
	}
	return months, nil
}
