package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Jurisdiction *string   `json:"jurisdiction,omitempty"`
	BarNumber    *string   `json:"bar_number,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type CaseResponse struct {
	ID          uuid.UUID `json:"id"`
	ClientID    uuid.UUID `json:"client_id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CaseWithQuotesResponse struct {
	CaseResponse
	QuotesCount int             `json:"quotes_count"`
	Quotes      []QuoteResponse `json:"quotes,omitempty"`
	Files       []FileResponse  `json:"files,omitempty"`
}

type MarketplaceCaseResponse struct {
	ID           uuid.UUID      `json:"id"`
	Title        string         `json:"title"`
	Category     string         `json:"category"`
	Description  string         `json:"description"`
	CreatedAt    time.Time      `json:"created_at"`
	Status       string         `json:"status,omitempty"`
	Files        []FileResponse `json:"files,omitempty"`
	HasSubmitted bool           `json:"has_submitted"`
}

type QuoteResponse struct {
	ID           uuid.UUID       `json:"id"`
	CaseID       uuid.UUID       `json:"case_id"`
	LawyerID     uuid.UUID       `json:"lawyer_id"`
	Amount       decimal.Decimal `json:"amount"`
	ExpectedDays int             `json:"expected_days"`
	Note         string          `json:"note"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	LawyerName   *string         `json:"lawyer_name,omitempty"`
	CaseTitle    *string         `json:"case_title,omitempty"`
}

type FileResponse struct {
	ID          uuid.UUID `json:"id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	CreatedAt   time.Time `json:"created_at"`
	DownloadURL *string   `json:"download_url,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

type PaymentIntentResponse struct {
	PaymentIntentID string `json:"payment_intent_id"`
	PaymentLinkURL  string `json:"payment_link_url"`
}
