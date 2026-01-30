package service

import (
	"context"
	"fmt"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CaseService struct {
	repo repository.Repository
}

func NewCaseService(repo repository.Repository) *CaseService {
	return &CaseService{
		repo: repo,
	}
}

func (s *CaseService) CreateCase(ctx context.Context, clientID uuid.UUID, req dto.CreateCaseRequest) (*dto.CaseResponse, error) {
	caseRecord, err := s.repo.CreateCase(ctx, &repository.CreateCaseParams{
		ClientID:    clientID,
		Title:       req.Title,
		Category:    req.Category,
		Description: req.Description,
		Status:      "open",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create case: %w", err)
	}

	return &dto.CaseResponse{
		ID:          caseRecord.ID,
		ClientID:    caseRecord.ClientID,
		Title:       caseRecord.Title,
		Category:    caseRecord.Category,
		Description: caseRecord.Description,
		Status:      caseRecord.Status,
		CreatedAt:   utils.PgtypeTimeToTime(caseRecord.CreatedAt),
		UpdatedAt:   utils.PgtypeTimeToTime(caseRecord.UpdatedAt),
	}, nil
}

func (s *CaseService) GetCasesByClientID(ctx context.Context, clientID uuid.UUID, page, pageSize int) ([]dto.CaseWithQuotesResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	cases, err := s.repo.GetCasesByClientID(ctx, &repository.GetCasesByClientIDParams{
		ClientID: clientID,
		Limit:    int32(pageSize),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cases: %w", err)
	}

	total, err := s.repo.CountCasesByClientID(ctx, clientID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cases: %w", err)
	}

	result := make([]dto.CaseWithQuotesResponse, 0, len(cases))
	for _, caseRecord := range cases {
		quotesCount, _ := s.repo.CountQuotesByCaseID(ctx, caseRecord.ID)
		
		caseResp := dto.CaseWithQuotesResponse{
			CaseResponse: dto.CaseResponse{
				ID:          caseRecord.ID,
				ClientID:    caseRecord.ClientID,
				Title:       caseRecord.Title,
				Category:    caseRecord.Category,
				Description: caseRecord.Description,
				Status:      caseRecord.Status,
				CreatedAt:   utils.PgtypeTimeToTime(caseRecord.CreatedAt),
				UpdatedAt:   utils.PgtypeTimeToTime(caseRecord.UpdatedAt),
			},
			QuotesCount: int(quotesCount),
		}
		result = append(result, caseResp)
	}

	return result, int64(total), nil
}

func getDecimalOrZero(d *decimal.Decimal) decimal.Decimal {
	if d == nil {
		return decimal.Zero
	}
	return *d
}


func (s *CaseService) GetCaseByID(ctx context.Context, caseID uuid.UUID, userID uuid.UUID, userRole string) (*dto.CaseWithQuotesResponse, error) {
	caseRecord, err := s.repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}


	if userRole == "client" && caseRecord.ClientID != userID {
		return nil, fmt.Errorf("unauthorized: you can only view your own cases")
	}


	quotes, err := s.repo.GetQuotesByCaseID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	quotesResp := make([]dto.QuoteResponse, 0, len(quotes))
	for _, quote := range quotes {
		quotesResp = append(quotesResp, dto.QuoteResponse{
			ID:           quote.ID,
			CaseID:       quote.CaseID,
			LawyerID:     quote.LawyerID,
			Amount:       getDecimalOrZero(utils.PgtypeNumericToDecimal(quote.Amount)),
			ExpectedDays: int(quote.ExpectedDays),
			Note:         utils.GetStringOrEmpty(utils.GetNullableString(quote.Note)),
			Status:       quote.Status,
			CreatedAt:    utils.PgtypeTimeToTime(quote.CreatedAt),
			UpdatedAt:    utils.PgtypeTimeToTime(quote.UpdatedAt),
			LawyerName:   utils.GetNullableString(quote.LawyerName),
		})
	}


	files := []dto.FileResponse{}
	if userRole == "client" || (userRole == "lawyer" && caseRecord.Status == "engaged") {
		caseFiles, _ := s.repo.GetCaseFilesByCaseID(ctx, caseID)
		for _, file := range caseFiles {
			files = append(files, dto.FileResponse{
				ID:       file.ID,
				FileName: file.FileName,
				FileSize: file.FileSize,
				MimeType: file.MimeType,
				CreatedAt: utils.PgtypeTimeToTime(file.CreatedAt),
			})
		}
	}

	return &dto.CaseWithQuotesResponse{
		CaseResponse: dto.CaseResponse{
			ID:          caseRecord.ID,
			ClientID:    caseRecord.ClientID,
			Title:       caseRecord.Title,
			Category:    caseRecord.Category,
			Description: caseRecord.Description,
			Status:      caseRecord.Status,
			CreatedAt:   utils.PgtypeTimeToTime(caseRecord.CreatedAt),
			UpdatedAt:   utils.PgtypeTimeToTime(caseRecord.UpdatedAt),
		},
		QuotesCount: len(quotesResp),
		Quotes:      quotesResp,
		Files:       files,
	}, nil
}

