




package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const CreatePayment = `-- name: CreatePayment :one
INSERT INTO payments (quote_id, stripe_payment_intent_id, amount, status)
VALUES ($1, $2, $3, $4)
RETURNING id, quote_id, stripe_payment_intent_id, amount, status, created_at, updated_at
`

type CreatePaymentParams struct {
	QuoteID               uuid.UUID      `json:"quote_id"`
	StripePaymentIntentID string         `json:"stripe_payment_intent_id"`
	Amount                pgtype.Numeric `json:"amount"`
	Status                string         `json:"status"`
}

func (q *Queries) CreatePayment(ctx context.Context, arg *CreatePaymentParams) (*Payment, error) {
	row := q.db.QueryRow(ctx, CreatePayment,
		arg.QuoteID,
		arg.StripePaymentIntentID,
		arg.Amount,
		arg.Status,
	)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.QuoteID,
		&i.StripePaymentIntentID,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetPaymentByID = `-- name: GetPaymentByID :one
SELECT id, quote_id, stripe_payment_intent_id, amount, status, created_at, updated_at FROM payments WHERE id = $1
`

func (q *Queries) GetPaymentByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	row := q.db.QueryRow(ctx, GetPaymentByID, id)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.QuoteID,
		&i.StripePaymentIntentID,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetPaymentByQuoteID = `-- name: GetPaymentByQuoteID :one
SELECT id, quote_id, stripe_payment_intent_id, amount, status, created_at, updated_at FROM payments WHERE quote_id = $1 LIMIT 1
`

func (q *Queries) GetPaymentByQuoteID(ctx context.Context, quoteID uuid.UUID) (*Payment, error) {
	row := q.db.QueryRow(ctx, GetPaymentByQuoteID, quoteID)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.QuoteID,
		&i.StripePaymentIntentID,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const GetPaymentByStripePaymentIntentID = `-- name: GetPaymentByStripePaymentIntentID :one
SELECT id, quote_id, stripe_payment_intent_id, amount, status, created_at, updated_at FROM payments WHERE stripe_payment_intent_id = $1
`

func (q *Queries) GetPaymentByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*Payment, error) {
	row := q.db.QueryRow(ctx, GetPaymentByStripePaymentIntentID, stripePaymentIntentID)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.QuoteID,
		&i.StripePaymentIntentID,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const UpdatePaymentStatus = `-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, quote_id, stripe_payment_intent_id, amount, status, created_at, updated_at
`

type UpdatePaymentStatusParams struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

func (q *Queries) UpdatePaymentStatus(ctx context.Context, arg *UpdatePaymentStatusParams) (*Payment, error) {
	row := q.db.QueryRow(ctx, UpdatePaymentStatus, arg.ID, arg.Status)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.QuoteID,
		&i.StripePaymentIntentID,
		&i.Amount,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
