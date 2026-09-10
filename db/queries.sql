-- Reference copy of the queries used by internal/repository. Kept here for
-- documentation/DBA review; the actual strings live next to the Go code that
-- runs them (internal/repository/sqlite_ficha_repository.go).

-- name: SaveFicha
INSERT INTO fichas (id, full_name, request_type, acs, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: FindFichas
-- Filters by name and/or month ("YYYY-MM") are applied conditionally in Go;
-- shown here with both applied.
SELECT id, full_name, request_type, acs, created_at
FROM fichas
WHERE full_name LIKE '%' || ? || '%'
  AND substr(created_at, 1, 7) = ?
ORDER BY created_at DESC;

-- name: FindFichaMonths
SELECT DISTINCT substr(created_at, 1, 7)
FROM fichas
ORDER BY 1 DESC;
