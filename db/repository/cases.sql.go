




package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const CountCasesByClientID = `-- name: CountCasesByClientID :one
SELECT COUNT(*) FROM cases WHERE client_id = $1
`

func (q *Queries) CountCasesByClientID(ctx context.Context, clientID uuid.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, CountCasesByClientID, clientID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const CountOpenCases = `-- name: CountOpenCases :one
SELECT COUNT(*) FROM cases
WHERE status = 'open'
  AND ($1::VARCHAR IS NULL OR $1 = '' OR category = $1)
  AND ($2::TIMESTAMPTZ IS NULL OR created_at >= $2)
`

type CountOpenCasesParams struct {
	Column1 string    `json:"column_1"`
	Column2 time.Time `json:"column_2"`
}

func (q *Queries) CountOpenCases(ctx context.Context, arg *CountOpenCasesParams) (int64, error) {
	row := q.db.QueryRow(ctx, CountOpenCases, arg.Column1, arg.Column2)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const CreateCase = `-- name: CreateCase :one
INSERT INTO cases (client_id, title, category, description, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, client_id, title, category, description, status, created_at, updated_at
`

type CreateCaseParams struct {
	ClientID    uuid.UUID `json:"client_id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
}

func (q *Queries) CreateCase(ctx context.Context, arg *CreateCaseParams) (*Case, error) {
	row := q.db.QueryRow(ctx, CreateCase,
		arg.ClientID,
		arg.Title,
		arg.Category,
		arg.Description,
		arg.Status,
	)
	var i Case
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.Title,
		&i.Category,
		&i.Description,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetCaseByID = `-- name: GetCaseByID :one
SELECT id, client_id, title, category, description, status, created_at, updated_at FROM cases WHERE id = $1
`

func (q *Queries) GetCaseByID(ctx context.Context, id uuid.UUID) (*Case, error) {
	row := q.db.QueryRow(ctx, GetCaseByID, id)
	var i Case
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.Title,
		&i.Category,
		&i.Description,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetCaseWithClient = `-- name: GetCaseWithClient :one
SELECT c.id, c.client_id, c.title, c.category, c.description, c.status, c.created_at, c.updated_at, u.name as client_name, u.email as client_email
FROM cases c
JOIN users u ON c.client_id = u.id
WHERE c.id = $1
`

type GetCaseWithClientRow struct {
	ID          uuid.UUID          `json:"id"`
	ClientID    uuid.UUID          `json:"client_id"`
	Title       string             `json:"title"`
	Category    string             `json:"category"`
	Description string             `json:"description"`
	Status      string             `json:"status"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
	ClientName  pgtype.Text        `json:"client_name"`
	ClientEmail string             `json:"client_email"`
}

func (q *Queries) GetCaseWithClient(ctx context.Context, id uuid.UUID) (*GetCaseWithClientRow, error) {
	row := q.db.QueryRow(ctx, GetCaseWithClient, id)
	var i GetCaseWithClientRow
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.Title,
		&i.Category,
		&i.Description,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ClientName,
		&i.ClientEmail,
	)
	return &i, err
}

const GetCasesByClientID = `-- name: GetCasesByClientID :many
SELECT id, client_id, title, category, description, status, created_at, updated_at FROM cases 
WHERE client_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`

type GetCasesByClientIDParams struct {
	ClientID uuid.UUID `json:"client_id"`
	Limit    int32     `json:"limit"`
	Offset   int32     `json:"offset"`
}

func (q *Queries) GetCasesByClientID(ctx context.Context, arg *GetCasesByClientIDParams) ([]*Case, error) {
	rows, err := q.db.Query(ctx, GetCasesByClientID, arg.ClientID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Case{}
	for rows.Next() {
		var i Case
		if err := rows.Scan(
			&i.ID,
			&i.ClientID,
			&i.Title,
			&i.Category,
			&i.Description,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const ListOpenCases = `-- name: ListOpenCases :many
SELECT c.id, c.client_id, c.title, c.category, c.description, c.status, c.created_at, c.updated_at, u.name as client_name
FROM cases c
JOIN users u ON c.client_id = u.id
WHERE c.status = 'open'
  AND ($1::VARCHAR IS NULL OR $1 = '' OR c.category = $1)
  AND ($2::TIMESTAMPTZ IS NULL OR c.created_at >= $2)
ORDER BY c.created_at DESC
LIMIT $3 OFFSET $4
`

type ListOpenCasesParams struct {
	Column1 string    `json:"column_1"`
	Column2 time.Time `json:"column_2"`
	Limit   int32     `json:"limit"`
	Offset  int32     `json:"offset"`
}

type ListOpenCasesRow struct {
	ID          uuid.UUID          `json:"id"`
	ClientID    uuid.UUID          `json:"client_id"`
	Title       string             `json:"title"`
	Category    string             `json:"category"`
	Description string             `json:"description"`
	Status      string             `json:"status"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
	ClientName  pgtype.Text        `json:"client_name"`
}

func (q *Queries) ListOpenCases(ctx context.Context, arg *ListOpenCasesParams) ([]*ListOpenCasesRow, error) {
	rows, err := q.db.Query(ctx, ListOpenCases,
		arg.Column1,
		arg.Column2,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*ListOpenCasesRow{}
	for rows.Next() {
		var i ListOpenCasesRow
		if err := rows.Scan(
			&i.ID,
			&i.ClientID,
			&i.Title,
			&i.Category,
			&i.Description,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ClientName,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const UpdateCaseStatus = `-- name: UpdateCaseStatus :one
UPDATE cases
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, client_id, title, category, description, status, created_at, updated_at
`

type UpdateCaseStatusParams struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (q *Queries) UpdateCaseStatus(ctx context.Context, arg *UpdateCaseStatusParams) (*Case, error) {
	row := q.db.QueryRow(ctx, UpdateCaseStatus, arg.ID, arg.Status)
	var i Case
	err := row.Scan(
		&i.ID,
		&i.ClientID,
		&i.Title,
		&i.Category,
		&i.Description,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
