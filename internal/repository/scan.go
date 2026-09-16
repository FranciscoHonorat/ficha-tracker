package repository

import (
	"database/sql"
	"fmt"
	"time"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

const rfc3339 = time.RFC3339Nano

// scanFicha scans a row produced by a query that selects, in order:
// id, full_name, request_type, acs_id, phone, notified, created_at, and
// optionally a trailing joined acs name (acsName is empty when not selected).
func scanFicha(rows *sql.Rows, withACSName bool) (*domain.Ficha, error) {
	var (
		idStr, fullName, requestType, acsIDStr, phone, createdAtStr string
		notified                                                    bool
		acsName                                                     sql.NullString
	)

	dest := []any{&idStr, &fullName, &requestType, &acsIDStr, &phone, &notified, &createdAtStr}
	if withACSName {
		dest = append(dest, &acsName)
	}

	if err := rows.Scan(dest...); err != nil {
		return nil, fmt.Errorf("scanning ficha: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing ficha id: %w", err)
	}

	acsID, err := uuid.Parse(acsIDStr)
	if err != nil {
		return nil, fmt.Errorf("parsing ficha acs_id: %w", err)
	}

	createdAt, err := time.Parse(rfc3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parsing ficha created_at: %w", err)
	}

	return &domain.Ficha{
		ID:          id,
		FullName:    fullName,
		RequestType: requestType,
		ACSID:       acsID,
		ACSName:     acsName.String,
		Phone:       phone,
		Notified:    notified,
		CreatedAt:   createdAt,
	}, nil
}

func scanFichas(rows *sql.Rows, withACSName bool) ([]*domain.Ficha, error) {
	var fichas []*domain.Ficha
	for rows.Next() {
		ficha, err := scanFicha(rows, withACSName)
		if err != nil {
			return nil, err
		}
		fichas = append(fichas, ficha)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading fichas: %w", err)
	}
	return fichas, nil
}

func scanACS(rows *sql.Rows) (*domain.ACS, error) {
	var (
		idStr, name, phone, createdAtStr string
	)

	if err := rows.Scan(&idStr, &name, &phone, &createdAtStr); err != nil {
		return nil, fmt.Errorf("scanning acs: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing acs id: %w", err)
	}

	createdAt, err := time.Parse(rfc3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parsing acs created_at: %w", err)
	}

	return &domain.ACS{
		ID:        id,
		Name:      name,
		Phone:     phone,
		CreatedAt: createdAt,
	}, nil
}

func scanACSList(rows *sql.Rows) ([]*domain.ACS, error) {
	var list []*domain.ACS
	for rows.Next() {
		acs, err := scanACS(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, acs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading acs: %w", err)
	}
	return list, nil
}
