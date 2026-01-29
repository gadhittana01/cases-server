package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-app-server/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type QuoteService struct {
	repo repository.Repository
}

func NewQuoteService(repo repository.Repository) *QuoteService {
	return &QuoteService{
		repo: repo,
	}
}

func (s *QuoteService) CreateQuote(ctx context.Context, caseID, lawyerID uuid.UUID, req dto.SubmitQuoteRequest) (*dto.QuoteResponse, error) {
	// Verify case exists and is open
	caseRecord, err := s.repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}
	if caseRecord.Status != "open" {
		return nil, fmt.Errorf("case is not open for quotes")
	}

	// Check if quote already exists - prevent duplicate submission
	existingQuote, err := s.repo.GetQuoteByCaseAndLawyer(ctx, &repository.GetQuoteByCaseAndLawyerParams{
		CaseID:   caseID,
		LawyerID: lawyerID,
	})
	if err == nil && existingQuote != nil {
		return nil, fmt.Errorf("you have already submitted a quote for this case. Use update endpoint to modify it")
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing quote: %w", err)
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	// Create new quote
	quote, err := s.repo.CreateQuote(ctx, &repository.CreateQuoteParams{
		CaseID:       caseID,
		LawyerID:     lawyerID,
		Amount:       utils.DecimalToPgtypeNumeric(amount),
		ExpectedDays: int32(req.ExpectedDays),
		Note:         utils.ToPgtypeText(&req.Note),
		Status:       "proposed",
	})
	if err != nil {
		// Check for unique constraint violation
		if err.Error() != "" {
			return nil, fmt.Errorf("failed to create quote: %w", err)
		}
		return nil, fmt.Errorf("you have already submitted a quote for this case")
	}

	return s.quoteToResponse(quote), nil
}

func (s *QuoteService) UpdateQuote(ctx context.Context, caseID, lawyerID uuid.UUID, req dto.SubmitQuoteRequest) (*dto.QuoteResponse, error) {
	// Verify case exists and is open
	caseRecord, err := s.repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}
	if caseRecord.Status != "open" {
		return nil, fmt.Errorf("case is not open for quotes")
	}

	// Get existing quote
	existingQuote, err := s.repo.GetQuoteByCaseAndLawyer(ctx, &repository.GetQuoteByCaseAndLawyerParams{
		CaseID:   caseID,
		LawyerID: lawyerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("quote not found. Please create a quote first")
		}
		return nil, fmt.Errorf("failed to get existing quote: %w", err)
	}

	// Prevent updating quotes that are already accepted
	if existingQuote.Status == "accepted" {
		return nil, fmt.Errorf("quote already accepted, cannot update")
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	// Update existing quote (UpdateQuote resets status to "proposed" if it was "rejected")
	quote, err := s.repo.UpdateQuote(ctx, &repository.UpdateQuoteParams{
		ID:           existingQuote.ID,
		Amount:       utils.DecimalToPgtypeNumeric(amount),
		ExpectedDays: int32(req.ExpectedDays),
		Note:         utils.ToPgtypeText(&req.Note),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	return s.quoteToResponse(quote), nil
}

func (s *QuoteService) quoteToResponse(quote *repository.Quote) *dto.QuoteResponse {
	amountDecimal := utils.PgtypeNumericToDecimal(quote.Amount)
	if amountDecimal == nil {
		amountDecimal = &decimal.Zero
	}

	return &dto.QuoteResponse{
		ID:           quote.ID,
		CaseID:       quote.CaseID,
		LawyerID:     quote.LawyerID,
		Amount:       *amountDecimal,
		ExpectedDays: int(quote.ExpectedDays),
		Note:         utils.GetStringOrEmpty(utils.GetNullableString(quote.Note)),
		Status:       quote.Status,
		CreatedAt:    utils.PgtypeTimeToTime(quote.CreatedAt),
		UpdatedAt:    utils.PgtypeTimeToTime(quote.UpdatedAt),
	}
}

func (s *QuoteService) GetQuoteByCaseAndLawyer(ctx context.Context, caseID, lawyerID uuid.UUID) (*dto.QuoteResponse, error) {
	quote, err := s.repo.GetQuoteByCaseAndLawyer(ctx, &repository.GetQuoteByCaseAndLawyerParams{
		CaseID:   caseID,
		LawyerID: lawyerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No quote exists, return nil (not an error)
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	return s.quoteToResponse(quote), nil
}

func (s *QuoteService) GetQuotesByLawyerID(ctx context.Context, lawyerID uuid.UUID, status string, page, pageSize int) ([]dto.QuoteResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	statusFilter := ""
	if status != "" {
		statusFilter = status
	}

	quotes, err := s.repo.GetQuotesByLawyerID(ctx, &repository.GetQuotesByLawyerIDParams{
		LawyerID: lawyerID,
		Column2:  statusFilter,
		Limit:    int32(pageSize),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get quotes: %w", err)
	}

	total, err := s.repo.CountQuotesByLawyerID(ctx, &repository.CountQuotesByLawyerIDParams{
		LawyerID: lawyerID,
		Column2:  statusFilter,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count quotes: %w", err)
	}

	result := make([]dto.QuoteResponse, 0, len(quotes))
	for _, quote := range quotes {
		amountDecimal := utils.PgtypeNumericToDecimal(quote.Amount)
		if amountDecimal == nil {
			amountDecimal = &decimal.Zero
		}

		result = append(result, dto.QuoteResponse{
			ID:           quote.ID,
			CaseID:       quote.CaseID,
			LawyerID:     quote.LawyerID,
			Amount:       *amountDecimal,
			ExpectedDays: int(quote.ExpectedDays),
			Note:         utils.GetStringOrEmpty(utils.GetNullableString(quote.Note)),
			Status:       quote.Status,
			CreatedAt:    utils.PgtypeTimeToTime(quote.CreatedAt),
			UpdatedAt:    utils.PgtypeTimeToTime(quote.UpdatedAt),
			CaseTitle:    &quote.CaseTitle,
		})
	}

	return result, int64(total), nil
}


