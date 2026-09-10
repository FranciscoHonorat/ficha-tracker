package repository

import (
	"database/sql"
	"fmt"
	"time"

	"ficha-tracker/internal/domain"

	"github.com/google/uuid"
)

const rfc3339 = time.RFC3339Nano

func scanFicha(rows *sql.Rows) (*domain.Ficha, error) {
	var (
		idStr, fullName, requestType, acs, createdAtStr string
	)

	if err := rows.Scan(&idStr, &fullName, &requestType, &acs, &createdAtStr); err != nil {
		return nil, fmt.Errorf("scanning ficha: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parsing ficha id: %w", err)
	}

	createdAt, err := time.Parse(rfc3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parsing ficha created_at: %w", err)
	}

	return &domain.Ficha{
		ID:          id,
		FullName:    fullName,
		RequestType: requestType,
		ACS:         acs,
		CreatedAt:   createdAt,
	}, nil
}
