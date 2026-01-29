-- name: CreatePayment :one
INSERT INTO payments (quote_id, stripe_payment_intent_id, amount, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPaymentByID :one
SELECT * FROM payments WHERE id = $1;

-- name: GetPaymentByStripePaymentIntentID :one
SELECT * FROM payments WHERE stripe_payment_intent_id = $1;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetPaymentByQuoteID :one
SELECT * FROM payments WHERE quote_id = $1 LIMIT 1;
