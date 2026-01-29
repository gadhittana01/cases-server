-- name: CreateCase :one
INSERT INTO cases (client_id, title, category, description, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCaseByID :one
SELECT * FROM cases WHERE id = $1;

-- name: GetCasesByClientID :many
SELECT * FROM cases 
WHERE client_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCasesByClientID :one
SELECT COUNT(*) FROM cases WHERE client_id = $1;

-- name: UpdateCaseStatus :one
UPDATE cases
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListOpenCases :many
SELECT c.*, u.name as client_name
FROM cases c
JOIN users u ON c.client_id = u.id
WHERE c.status = 'open'
  AND ($1::VARCHAR IS NULL OR $1 = '' OR c.category = $1)
  AND ($2::TIMESTAMPTZ IS NULL OR c.created_at >= $2)
ORDER BY c.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountOpenCases :one
SELECT COUNT(*) FROM cases
WHERE status = 'open'
  AND ($1::VARCHAR IS NULL OR $1 = '' OR category = $1)
  AND ($2::TIMESTAMPTZ IS NULL OR created_at >= $2);

-- name: GetCaseWithClient :one
SELECT c.*, u.name as client_name, u.email as client_email
FROM cases c
JOIN users u ON c.client_id = u.id
WHERE c.id = $1;
