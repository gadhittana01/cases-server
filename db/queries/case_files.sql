-- name: CreateCaseFile :one
INSERT INTO case_files (case_id, file_name, file_path, file_size, mime_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCaseFileByID :one
SELECT * FROM case_files WHERE id = $1;

-- name: GetCaseFilesByCaseID :many
SELECT * FROM case_files 
WHERE case_id = $1
ORDER BY created_at ASC;

-- name: CountCaseFilesByCaseID :one
SELECT COUNT(*) FROM case_files WHERE case_id = $1;

-- name: DeleteCaseFile :exec
DELETE FROM case_files WHERE id = $1;
