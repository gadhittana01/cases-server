




package repository

import (
	"context"

	"github.com/google/uuid"
)

const CountCaseFilesByCaseID = `-- name: CountCaseFilesByCaseID :one
SELECT COUNT(*) FROM case_files WHERE case_id = $1
`

func (q *Queries) CountCaseFilesByCaseID(ctx context.Context, caseID uuid.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, CountCaseFilesByCaseID, caseID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const CreateCaseFile = `-- name: CreateCaseFile :one
INSERT INTO case_files (case_id, file_name, file_path, file_size, mime_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, case_id, file_name, file_path, file_size, mime_type, created_at
`

type CreateCaseFileParams struct {
	CaseID   uuid.UUID `json:"case_id"`
	FileName string    `json:"file_name"`
	FilePath string    `json:"file_path"`
	FileSize int64     `json:"file_size"`
	MimeType string    `json:"mime_type"`
}

func (q *Queries) CreateCaseFile(ctx context.Context, arg *CreateCaseFileParams) (*CaseFile, error) {
	row := q.db.QueryRow(ctx, CreateCaseFile,
		arg.CaseID,
		arg.FileName,
		arg.FilePath,
		arg.FileSize,
		arg.MimeType,
	)
	var i CaseFile
	err := row.Scan(
		&i.ID,
		&i.CaseID,
		&i.FileName,
		&i.FilePath,
		&i.FileSize,
		&i.MimeType,
		&i.CreatedAt,
	)
	return &i, err
}

const DeleteCaseFile = `-- name: DeleteCaseFile :exec
DELETE FROM case_files WHERE id = $1
`

func (q *Queries) DeleteCaseFile(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, DeleteCaseFile, id)
	return err
}

const GetCaseFileByID = `-- name: GetCaseFileByID :one
SELECT id, case_id, file_name, file_path, file_size, mime_type, created_at FROM case_files WHERE id = $1
`

func (q *Queries) GetCaseFileByID(ctx context.Context, id uuid.UUID) (*CaseFile, error) {
	row := q.db.QueryRow(ctx, GetCaseFileByID, id)
	var i CaseFile
	err := row.Scan(
		&i.ID,
		&i.CaseID,
		&i.FileName,
		&i.FilePath,
		&i.FileSize,
		&i.MimeType,
		&i.CreatedAt,
	)
	return &i, err
}

const GetCaseFilesByCaseID = `-- name: GetCaseFilesByCaseID :many
SELECT id, case_id, file_name, file_path, file_size, mime_type, created_at FROM case_files 
WHERE case_id = $1
ORDER BY created_at ASC
`

func (q *Queries) GetCaseFilesByCaseID(ctx context.Context, caseID uuid.UUID) ([]*CaseFile, error) {
	rows, err := q.db.Query(ctx, GetCaseFilesByCaseID, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*CaseFile{}
	for rows.Next() {
		var i CaseFile
		if err := rows.Scan(
			&i.ID,
			&i.CaseID,
			&i.FileName,
			&i.FilePath,
			&i.FileSize,
			&i.MimeType,
			&i.CreatedAt,
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
