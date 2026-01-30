




package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const CreateUser = `-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, role, jurisdiction, bar_number)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, password_hash, name, role, jurisdiction, bar_number, created_at, updated_at
`

type CreateUserParams struct {
	Email        string      `json:"email"`
	PasswordHash string      `json:"password_hash"`
	Name         pgtype.Text `json:"name"`
	Role         string      `json:"role"`
	Jurisdiction pgtype.Text `json:"jurisdiction"`
	BarNumber    pgtype.Text `json:"bar_number"`
}

func (q *Queries) CreateUser(ctx context.Context, arg *CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, CreateUser,
		arg.Email,
		arg.PasswordHash,
		arg.Name,
		arg.Role,
		arg.Jurisdiction,
		arg.BarNumber,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.Role,
		&i.Jurisdiction,
		&i.BarNumber,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, role, jurisdiction, bar_number, created_at, updated_at FROM users WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := q.db.QueryRow(ctx, GetUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.Role,
		&i.Jurisdiction,
		&i.BarNumber,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetUserByID = `-- name: GetUserByID :one
SELECT id, email, password_hash, name, role, jurisdiction, bar_number, created_at, updated_at FROM users WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := q.db.QueryRow(ctx, GetUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.Role,
		&i.Jurisdiction,
		&i.BarNumber,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const UpdateUser = `-- name: UpdateUser :one
UPDATE users
SET name = COALESCE($2, name),
    jurisdiction = COALESCE($3, jurisdiction),
    bar_number = COALESCE($4, bar_number),
    updated_at = NOW()
WHERE id = $1
RETURNING id, email, password_hash, name, role, jurisdiction, bar_number, created_at, updated_at
`

type UpdateUserParams struct {
	ID           uuid.UUID   `json:"id"`
	Name         pgtype.Text `json:"name"`
	Jurisdiction pgtype.Text `json:"jurisdiction"`
	BarNumber    pgtype.Text `json:"bar_number"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg *UpdateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, UpdateUser,
		arg.ID,
		arg.Name,
		arg.Jurisdiction,
		arg.BarNumber,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.Role,
		&i.Jurisdiction,
		&i.BarNumber,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
