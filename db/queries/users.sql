-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, role, jurisdiction, bar_number)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET name = COALESCE($2, name),
    jurisdiction = COALESCE($3, jurisdiction),
    bar_number = COALESCE($4, bar_number),
    updated_at = NOW()
WHERE id = $1
RETURNING *;
