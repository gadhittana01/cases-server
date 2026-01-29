package dto

type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Name        string `json:"name"`
	Role        string `json:"role" binding:"required,oneof=client lawyer"`
	Jurisdiction string `json:"jurisdiction"`
	BarNumber   string `json:"bar_number"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateCaseRequest struct {
	Title       string `json:"title" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type SubmitQuoteRequest struct {
	Amount       string `json:"amount" binding:"required"`
	ExpectedDays int    `json:"expected_days" binding:"required,min=1"`
	Note         string `json:"note"`
}

type AcceptQuoteRequest struct {
	QuoteID string `json:"quote_id" binding:"required,uuid"`
}

type MarketplaceFilters struct {
	Category    string `form:"category"`
	CreatedSince string `form:"created_since"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

type MyQuotesFilters struct {
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
