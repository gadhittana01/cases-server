-- name: CreateQuote :one
INSERT INTO quotes (case_id, lawyer_id, amount, expected_days, note, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateQuote :one
UPDATE quotes
SET amount = $2, expected_days = $3, note = $4, status = 'proposed', updated_at = NOW()
WHERE id = $1 AND status != 'accepted'
RETURNING *;

-- name: GetQuoteByID :one
SELECT * FROM quotes WHERE id = $1;

-- name: GetQuoteByCaseAndLawyer :one
SELECT * FROM quotes 
WHERE case_id = $1 AND lawyer_id = $2;

-- name: GetQuotesByCaseID :many
SELECT q.*, u.name as lawyer_name, u.jurisdiction as lawyer_jurisdiction
FROM quotes q
JOIN users u ON q.lawyer_id = u.id
WHERE q.case_id = $1
ORDER BY q.created_at ASC;

-- name: CountQuotesByCaseID :one
SELECT COUNT(*) FROM quotes WHERE case_id = $1;

-- name: GetQuotesByLawyerID :many
SELECT q.*, c.title as case_title, c.category as case_category, c.status as case_status
FROM quotes q
JOIN cases c ON q.case_id = c.id
WHERE q.lawyer_id = $1
  AND ($2::VARCHAR IS NULL OR $2::VARCHAR = '' OR q.status = $2)
ORDER BY q.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountQuotesByLawyerID :one
SELECT COUNT(*) FROM quotes
WHERE lawyer_id = $1
  AND ($2::VARCHAR IS NULL OR $2::VARCHAR = '' OR status = $2);

-- name: AcceptQuote :one
UPDATE quotes
SET status = 'accepted', updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RejectOtherQuotes :many
UPDATE quotes
SET status = 'rejected', updated_at = NOW()
WHERE case_id = $1 AND id != $2 AND status = 'proposed'
RETURNING *;

-- name: GetAcceptedQuoteByCaseID :one
SELECT * FROM quotes 
WHERE case_id = $1 AND status = 'accepted'
LIMIT 1;
