package handler

import (
	"net/http"
	"strconv"

	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuoteHandler struct {
	quoteService *service.QuoteService
}

func NewQuoteHandler(quoteService *service.QuoteService) *QuoteHandler {
	return &QuoteHandler{
		quoteService: quoteService,
	}
}

func (h *QuoteHandler) CreateQuote(c *gin.Context) {
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

	var req dto.SubmitQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.quoteService.CreateQuote(c.Request.Context(), caseID, lawyerID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *QuoteHandler) UpdateQuote(c *gin.Context) {
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

	var req dto.SubmitQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.quoteService.UpdateQuote(c.Request.Context(), caseID, lawyerID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *QuoteHandler) GetMyQuoteForCase(c *gin.Context) {
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

	quote, err := h.quoteService.GetQuoteByCaseAndLawyer(c.Request.Context(), caseID, lawyerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if quote == nil {
		c.JSON(http.StatusOK, gin.H{"quote": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quote": quote})
}

func (h *QuoteHandler) GetMyQuotes(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	lawyerID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	
	// Support both 'limit' and 'page_size' query parameters
	pageSizeStr := c.Query("limit")
	if pageSizeStr == "" {
		pageSizeStr = c.DefaultQuery("page_size", "10")
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)

	quotes, total, err := h.quoteService.GetQuotesByLawyerID(c.Request.Context(), lawyerID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       quotes,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}
