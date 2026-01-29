package handler

import (
	"net/http"
	"strconv"

	"github.com/gadhittana01/cases-app-server/dto"
	"github.com/gadhittana01/cases-app-server/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CaseHandler struct {
	caseService *service.CaseService
	fileService *service.FileService
}

func NewCaseHandler(caseService *service.CaseService, fileService *service.FileService) *CaseHandler {
	return &CaseHandler{
		caseService: caseService,
		fileService: fileService,
	}
}

func (h *CaseHandler) CreateCase(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	clientID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req dto.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.caseService.CreateCase(c.Request.Context(), clientID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CaseHandler) GetMyCases(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	clientID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	cases, total, err := h.caseService.GetCasesByClientID(c.Request.Context(), clientID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       cases,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *CaseHandler) GetCaseByID(c *gin.Context) {
	caseIDStr := c.Param("id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	role, _ := c.Get("role")
	roleStr := role.(string)

	response, err := h.caseService.GetCaseByID(c.Request.Context(), caseID, userUUID, roleStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CaseHandler) UploadFile(c *gin.Context) {
	caseIDStr := c.Param("id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	// Authorization is handled in the file service

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	response, err := h.fileService.UploadCaseFile(c.Request.Context(), caseID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
