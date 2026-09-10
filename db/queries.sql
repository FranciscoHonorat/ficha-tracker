-- Reference copy of the queries used by internal/repository. Kept here for
-- documentation/DBA review; the actual strings live next to the Go code that
-- runs them (internal/repository/sqlite_ficha_repository.go).

-- name: SaveFicha
INSERT INTO fichas (id, full_name, request_type, acs, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: FindAllFichas
SELECT id, full_name, request_type, acs, created_at
FROM fichas
ORDER BY created_at DESC;

-- name: FindFichasByName
SELECT id, full_name, request_type, acs, created_at
FROM fichas
WHERE full_name LIKE '%' || ? || '%'
ORDER BY created_at DESC;
