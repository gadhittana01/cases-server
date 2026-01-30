package handler

import (
	"net/http"
	"strconv"

	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MarketplaceHandler struct {
	marketplaceService *service.MarketplaceService
}

func NewMarketplaceHandler(marketplaceService *service.MarketplaceService) *MarketplaceHandler {
	return &MarketplaceHandler{
		marketplaceService: marketplaceService,
	}
}

func (h *MarketplaceHandler) ListOpenCases(c *gin.Context) {
	var filters dto.MarketplaceFilters
	filters.Category = c.Query("category")
	filters.CreatedSince = c.Query("created_since")
	filters.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filters.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))

	cases, total, err := h.marketplaceService.ListOpenCases(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       cases,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *MarketplaceHandler) GetCaseForMarketplace(c *gin.Context) {
	caseIDStr := c.Param("id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}


	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	lawyerID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	response, err := h.marketplaceService.GetCaseForMarketplace(c.Request.Context(), caseID, &lawyerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
