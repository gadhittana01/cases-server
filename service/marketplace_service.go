package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gadhittana01/cases-app-server/db/repository"
	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-modules/utils"
	"github.com/google/uuid"
)

type MarketplaceService struct {
	repo repository.Repository
}

func NewMarketplaceService(repo repository.Repository) *MarketplaceService {
	return &MarketplaceService{
		repo: repo,
	}
}

func (s *MarketplaceService) ListOpenCases(ctx context.Context, filters dto.MarketplaceFilters) ([]dto.MarketplaceCaseResponse, int64, error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	categoryFilter := filters.Category

	var createdSinceFilter time.Time
	if filters.CreatedSince != "" {
		parsedTime, err := time.Parse(time.RFC3339, filters.CreatedSince)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid created_since format, use ISO 8601: %w", err)
		}
		createdSinceFilter = parsedTime
	}

	cases, err := s.repo.ListOpenCases(ctx, &repository.ListOpenCasesParams{
		Column1: categoryFilter,
		Column2: createdSinceFilter,
		Limit:   int32(pageSize),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list cases: %w", err)
	}

	total, err := s.repo.CountOpenCases(ctx, &repository.CountOpenCasesParams{
		Column1: categoryFilter,
		Column2: createdSinceFilter,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count cases: %w", err)
	}

	result := make([]dto.MarketplaceCaseResponse, 0, len(cases))
	for _, caseRecord := range cases {

		description := anonymizeDescription(caseRecord.Description)

		result = append(result, dto.MarketplaceCaseResponse{
			ID:          caseRecord.ID,
			Title:       caseRecord.Title,
			Category:    caseRecord.Category,
			Description: description,
			CreatedAt:   utils.PgtypeTimeToTime(caseRecord.CreatedAt),
		})
	}

	return result, int64(total), nil
}

func (s *MarketplaceService) GetCaseForMarketplace(ctx context.Context, caseID uuid.UUID, lawyerID *uuid.UUID) (*dto.MarketplaceCaseResponse, error) {
	caseRecord, err := s.repo.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}

	description := anonymizeDescription(caseRecord.Description)

	response := &dto.MarketplaceCaseResponse{
		ID:           caseRecord.ID,
		Title:        caseRecord.Title,
		Category:     caseRecord.Category,
		Description:  description,
		CreatedAt:    utils.PgtypeTimeToTime(caseRecord.CreatedAt),
		Status:       caseRecord.Status,
		HasSubmitted: false,
	}

	quote, err := s.repo.GetQuoteByCaseAndLawyer(ctx, &repository.GetQuoteByCaseAndLawyerParams{
		CaseID:   caseID,
		LawyerID: *lawyerID,
	})
	if err == nil && quote != nil {
		response.HasSubmitted = true
	}

	acceptedQuote, err := s.repo.GetAcceptedQuoteByCaseID(ctx, caseID)
	if err == nil && acceptedQuote != nil && acceptedQuote.LawyerID == *lawyerID {

		if caseRecord.Status == "engaged" {

			response.Description = caseRecord.Description

			caseFiles, err := s.repo.GetCaseFilesByCaseID(ctx, caseID)
			if err == nil {
				files := make([]dto.FileResponse, 0, len(caseFiles))
				for _, file := range caseFiles {
					files = append(files, dto.FileResponse{
						ID:        file.ID,
						FileName:  file.FileName,
						FileSize:  file.FileSize,
						MimeType:  file.MimeType,
						CreatedAt: utils.PgtypeTimeToTime(file.CreatedAt),
					})
				}
				response.Files = files
			}
		}
	}

	return response, nil
}

func anonymizeDescription(description string) string {

	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	description = emailRegex.ReplaceAllString(description, "[email redacted]")

	phoneRegex := regexp.MustCompile(`(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
	description = phoneRegex.ReplaceAllString(description, "[phone redacted]")

	description = strings.ReplaceAll(description, "@", "[at]")

	return description
}
